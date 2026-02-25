// WAL 追加日志 + 内存索引 + 定期压缩
// data.log：只追加写（append-only）
// data.compact：压缩时生成的新文件-临时文件
// 数据格式
//
//	crc32 | op(put/del) | keyLen | valLen | key | value
//
// 1) 启动恢复
//
//	进程启动时顺序扫描 data.log
//	构建内存索引：map[key] -> offset/len/状态
//
// 2) 写入流程
//
//	Set(k,v)：追加一条 put 记录
//	Delete(k)：追加一条 del（墓碑）记录
//	更新内存索引
//	根据策略 fsync：
//		强一致：每次写后 fsync
//		高性能：批量/定时 fsync
//
// 3) 读取流程
//
//	Get(k)：内存索引定位 -> data.log 偏移 -> 读取 value
//
// 4) 压缩（Compaction）
//
//	定期扫描 data.log，丢弃已删除或过期的键，生成新的 compact 文件
//	替换旧的 data.log，释放磁盘空间
//
// 5) 最小可用版本（MVP）
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"
)

type KVStore interface {
	// Get retrieves the value for a given key.
	// []byte：数据本身
	// bool：是否命中（found / not found）
	// error：是否发生异常（I/O、解码等）
	Get(key string) ([]byte, bool, error)

	// Set sets the value for a given key.
	Set(key string, value []byte) error

	// Delete removes the value for a given key.
	Delete(key string) error

	// ForEach iterates over all key-value pairs in the store.
	ForEach(fn func(key string, value []byte) error) error

	// Compact compacts the store to reclaim space.
	Compact() error

	// Close closes the store and releases any resources.
	Close() error
}

type FileKVStoreOptions struct {
	SyncOnWrite bool // Whether to sync to disk on every write.
}

type FileKVStore struct {
	mu   sync.RWMutex
	path string // Path to the store file.
	file *os.File
	// In-memory data store.
	data        map[string][]byte
	syncOnWrite bool
}

const (
	recordOpPut    byte = 1 // Put operation
	recordOpDelete byte = 2 // Delete operation
)

func NewFileKVStore(path string, options FileKVStoreOptions) (*FileKVStore, error) {
	if path == "" {
		return nil, errors.New("store path is empty")
	}

	// 创建存储目录，如果已存在就忽略
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	parsedData, lastGood, err := replayLog(path)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open store file: %w", err)
	}

	// 把文件截断到“最后一条有效记录”的位置，清掉崩溃导致的尾部脏数据/半条记录。
	if err := f.Truncate(lastGood); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("truncate store file: %w", err)
	}
	// 把文件指针移动到文件末尾，准备追加写入新记录
	if _, err := f.Seek(0, 2); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("seek store file: %w", err)
	}

	return &FileKVStore{
		path:        path,
		file:        f,
		data:        parsedData,
		syncOnWrite: options.SyncOnWrite,
	}, nil
}

func (s *FileKVStore) Get(key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return nil, false, nil
	}

	// 返回一份数据的副本，保证存储内部数据不被意外改写。
	out := make([]byte, len(v))
	copy(out, v)
	return out, true, nil
}

func (s *FileKVStore) Set(key string, value []byte) error {
	if key == "" {
		return errors.New("key is empty")
	}

	data := make([]byte, len(value))
	copy(data, value)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := writeRecord(s.file, recordOpPut, key, data); err != nil {
		return err
	}
	if s.syncOnWrite {
		if err := s.file.Sync(); err != nil {
			return fmt.Errorf("sync record: %w", err)
		}
	}
	s.data[key] = data
	return nil
}

func (s *FileKVStore) Delete(key string) error {
	if key == "" {
		return errors.New("key is empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 将该数据标记为删除，并追加写入日志
	if err := writeRecord(s.file, recordOpDelete, key, nil); err != nil {
		return err
	}
	if s.syncOnWrite {
		if err := s.file.Sync(); err != nil {
			return fmt.Errorf("sync delete record: %w", err)
		}
	}
	delete(s.data, key)
	return nil
}

func (s *FileKVStore) ForEach(fn func(key string, value []byte) error) error {
	if fn == nil {
		return errors.New("foreach function is nil")
	}

	s.mu.RLock()
	entries := make([]struct {
		k string
		v []byte
	}, 0, len(s.data))
	for k, v := range s.data {
		val := make([]byte, len(v))
		copy(val, v)
		entries = append(entries, struct {
			k string
			v []byte
		}{k: k, v: val})
	}
	s.mu.RUnlock()

	for _, e := range entries {
		if err := fn(e.k, e.v); err != nil {
			return err
		}
	}

	return nil
}

// Compact 压缩存储文件，移除已删除的数据，减少文件体积。
func (s *FileKVStore) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpPath := s.path + ".compact"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open compact file: %w", err)
	}

	// 将完整数据迁移到临时文件
	// 没有删除标记的垃圾数据
	for k, v := range s.data {
		if err := writeRecord(tmpFile, recordOpPut, k, v); err != nil {
			_ = tmpFile.Close()
			return err
		}
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync compact file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close compact file: %w", err)
	}

	if err := s.file.Close(); err != nil {
		return fmt.Errorf("close old store file: %w", err)
	}

	// 用临时文件替换旧的存储文件
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("replace compact file: %w", err)
	}

	newFile, err := os.OpenFile(s.path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("reopen compacted store file: %w", err)
	}
	if _, err := newFile.Seek(0, 2); err != nil {
		_ = newFile.Close()
		return fmt.Errorf("seek compacted store file: %w", err)
	}

	s.file = newFile
	return nil
}

func (s *FileKVStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		return nil
	}

	err := s.file.Close()
	s.file = nil
	return err
}

// replayLog replays the log file and returns the in-memory data, last good position, and error if any.
func replayLog(path string) (map[string][]byte, int64, error) {
	data := make(map[string][]byte)

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data, 0, nil
		}
		return nil, 0, fmt.Errorf("read store file: %w", err)
	}

	var (
		pos      int // Current position in the log
		lastGood int // 可恢复提交点
	)
	for {
		record, nextPos, ok := parseRecord(raw, pos)
		if !ok {
			break
		}

		// 如果: 坏记录后继续扫描/跳过未知 op, pos 和 lastGood 可能不同

		switch record.op {
		case recordOpPut:
			val := make([]byte, len(record.value))
			copy(val, record.value)
			data[record.key] = val
		case recordOpDelete:
			delete(data, record.key)
		}

		pos = nextPos
		lastGood = nextPos
	}

	return data, int64(lastGood), nil
}

type logRecord struct {
	// op is the operation type (put or delete).
	op    byte
	key   string
	value []byte
}

func parseRecord(raw []byte, pos int) (logRecord, int, bool) {
	// 记录数据格式
	// [4字节CRC][1字节op][4字节keyLen][4字节valueLen][key][value]
	const headerSize = 4 + 1 + 4 + 4
	if pos+headerSize > len(raw) {
		return logRecord{}, pos, false
	}

	expectedCRC := binary.BigEndian.Uint32(raw[pos : pos+4])      // → 取 CRC（4 字节）
	op := raw[pos+4]                                              // → 取操作类型（1 字节）
	keyLen := int(binary.BigEndian.Uint32(raw[pos+5 : pos+9]))    // → 取 key 长度（4 字节）
	valueLen := int(binary.BigEndian.Uint32(raw[pos+9 : pos+13])) // → 取 value 长度（4 字节）
	if keyLen < 0 || valueLen < 0 {
		return logRecord{}, pos, false
	}

	recordSize := headerSize + keyLen + valueLen
	if pos+recordSize > len(raw) {
		return logRecord{}, pos, false
	}

	keyStart := pos + headerSize
	valueStart := keyStart + keyLen
	checksumData := raw[pos+4 : pos+recordSize]
	actualCRC := crc32.ChecksumIEEE(checksumData) // crc32 检验数据的完整性
	if actualCRC != expectedCRC {
		return logRecord{}, pos, false
	}

	record := logRecord{
		op:    op,
		key:   string(raw[keyStart:valueStart]),
		value: raw[valueStart : valueStart+valueLen],
	}
	return record, pos + recordSize, true
}

func writeRecord(f *os.File, op byte, key string, value []byte) error {
	keyBytes := []byte(key)

	var payload bytes.Buffer
	payload.WriteByte(op)
	// 长度 写进二进制日志头
	if err := binary.Write(&payload, binary.BigEndian, uint32(len(keyBytes))); err != nil {
		return fmt.Errorf("encode key len: %w", err)
	}
	if err := binary.Write(&payload, binary.BigEndian, uint32(len(value))); err != nil {
		return fmt.Errorf("encode value len: %w", err)
	}
	payload.Write(keyBytes)
	payload.Write(value)

	crc := crc32.ChecksumIEEE(payload.Bytes())

	var record bytes.Buffer
	if err := binary.Write(&record, binary.BigEndian, crc); err != nil {
		return fmt.Errorf("encode crc: %w", err)
	}
	record.Write(payload.Bytes())

	if _, err := f.Write(record.Bytes()); err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	return nil
}
