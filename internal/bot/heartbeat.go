package bot

import (
	"bot/internal/pkg/timex"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"
)

const (
	attemptLimit = 3 // 连续失败多少次后才标记为异常，避免偶发的网络问题导致误判
)

type Health interface {
	Cfg() *HeartbeatConfig
	List() []int64 // 获取需要检测的ID列表
	Ping(idx int, id int64) error
	OnError(id int64, err error)
}

type HeartbeatConfig struct {
	HealthName   string        // 功能名称
	InitialDelay time.Duration // 初始延迟
	Interval     time.Duration // 检测间隔
	EmptyWait    time.Duration // 无活跃Bot时的等待时间
}

func runHealthCheck(ctx context.Context, health Health) {
	cfg := health.Cfg()
	logger := slog.With("health_name", cfg.HealthName)
	defer logger.Info("health check goroutine exiting")

	if timex.Sleep(ctx, cfg.InitialDelay) {
		logger.Info("health check initial wait interrupted, exiting")
		return // 初始等待被中断，直接退出
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("health check stopped, exiting")
			return
		default:
		}

		roundStartTime := time.Now()

		// 获取需要检测的bot列表（排除已禁用的和已关闭的）
		ids := health.List()
		if len(ids) == 0 {
			logger.Warn("no ready bots found during health check")
			if timex.Sleep(ctx, cfg.EmptyWait) {
				logger.Info("health check empty wait interrupted, exiting")
				return // 无活跃Bot时的等待被中断，直接退出
			}
			continue
		}

		// 随机打乱检测顺序，避免总是同一批Bot先被检测
		rand.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		// 计算每个 bot 的检测间隔，确保在 cfg.Interval 内完成所有检测
		checkInterval := cfg.Interval / time.Duration(len(ids))
		logger.Info("starting health check for ready bots", "bot_count", len(ids), "interval", checkInterval)

		// 轮询检测每个 bot 的健康状态
		for i, id := range ids {
			select {
			case <-ctx.Done():
				logger.Info("health check interrupted during bot checks, exiting", "checked_bots", i)
				return
			default:
			}

			idx := i + 1 // 本轮的第几个 bot

			// 检测逻辑
			err := health.Ping(idx, id)
			if err != nil {
				health.OnError(id, err)
				logger.Error("health check failed for bot", "bot_id", id, "error", err)
			}

			// 最后一个 bot 不需要等待检测间隔
			if i < len(ids)-1 {
				if timex.Sleep(ctx, checkInterval) {
					logger.Info("health check interrupted during bot checks, exiting", "checked_bots", idx)
					return // 检测间隔等待被中断，退出
				}
			}
		}

		elapsed := time.Since(roundStartTime)
		// 如果本轮检测提前完成，等待剩余时间，确保每轮间隔固定
		if elapsed < cfg.Interval {
			waitTime := cfg.Interval - elapsed
			if timex.Sleep(ctx, waitTime) {
				logger.Info("health check interval wait interrupted, exiting", "elapsed", elapsed, "wait_time", waitTime)
				return // 间隔等待被中断，退出
			}
		}
	}
}

const alertTmpl = `
🚨 Telegram 健康告警
{{if .AtUsernames}}{{.AtUsernames}}
{{end}}━━━━━━━━━━━━━━
🕒 检测时间：{{.Time}}
📌 告警对象：{{.Title}}
{{if .Id}}🆔 对象 ID：{{.Id}}
{{end}}{{if .Link}}🔗 访问链接：{{.Link}}
{{end}}📊 当前状态：{{.Status}}
❗ 异常详情：{{.Error}}
`

var templateAlert *template.Template

func init() {
	templateAlert = template.Must(template.New("alert").Parse(alertTmpl))
}

type AlertData struct {
	AtUsernames string // @username1 @username2
	Time        string // 2006/1/2 15:04:05
	// 1. BOT：@bot (https://t.me/bear_win_familytwo_bot)_xxx
	// 2. Channel：@xxx (https://t.me/bear_win_familytwo_bot)_xxx
	Title    string
	Username string
	Id       int64
	Link     string
	Status   string
	Error    string
}

func (a *AlertData) makeAtUsernames() {
	if a.AtUsernames == "" {
		return
	}
	a.AtUsernames = "@" + strings.ReplaceAll(a.AtUsernames, ",", " @")
}

func (a *AlertData) title(isBot bool) {
	if isBot {
		a.Title = fmt.Sprintf("Bot: @%s (%s/%s)", a.Username, DomainMe, a.Username)
	} else {
		a.Title = fmt.Sprintf("Channel: @%s (%s)", a.Username, a.Link)
	}
}

func renderAlert(data *AlertData) string {
	if data.Time == "" {
		data.Time = time.Now().Format(time.DateTime)
	}
	var sb strings.Builder
	if err := templateAlert.Execute(&sb, data); err != nil {
		slog.Error("renderAlert execute failed", "data", data, "error", err)
		return ""
	}
	return sb.String()
}

// sendChannelAlert 发送频道不可用告警
func (mgr *Manager) sendAlert(chatId int64, alertData *AlertData) {
	alertLarkMsg := renderAlert(alertData)

	alertBot := mgr.GetAlertBot()
	if alertBot == nil {
		slog.Error("sendAlert failed, no alert bot available", "alert_msg", alertLarkMsg)
		return
	}

	alertData.AtUsernames = "@xxx" // TODO: 配置可告警的用户名列表
	alertData.makeAtUsernames()
	alertMsg := renderAlert(alertData)
	if _, sendErr := alertBot.SendMsg(context.Background(), &SendMsgReq{
		ChatId:  chatId,
		Caption: alertMsg,
	}); sendErr != nil {
		slog.Error("[Alert] send warn bot alert failed", "error", sendErr)
	}
}
