CREATE TABLE telegram_bot (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID，自增',
    bot_tg_id BIGINT NOT NULL COMMENT '机器人在Telegram的ID',
    username VARCHAR(255) DEFAULT '' COMMENT '机器人用户名',
    token VARCHAR(255) NOT NULL COMMENT '机器人Token',
    webhook_secret VARCHAR(255) DEFAULT '' COMMENT 'Webhook认证密钥，对应Telegram的X-Telegram-Bot-Api-Secret-Token头',
    `owner` VARCHAR(255) DEFAULT '' COMMENT '机器人拥有者的Telegram用户名',
    `type` INT DEFAULT 0 COMMENT '机器人类型',
    health_group_id BIGINT DEFAULT 0 COMMENT '健康检查群组ID',
    `status` INT DEFAULT 1 COMMENT '状态：1.启用 2.禁用 3.封禁',
    created_at BIGINT DEFAULT 0 COMMENT '创建时间戳',
    updated_at BIGINT DEFAULT 0 COMMENT '更新时间戳',
    UNIQUE KEY uk_bot_tg_id (bot_tg_id),
    UNIQUE KEY uk_bot_username (username(191)),
    KEY idx_bot_owner (`owner`(191)),
    KEY idx_bot_status (`status`),
    KEY idx_bot_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='telegram机器人表';

CREATE TABLE telegram_channel (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID，自增',
    tg_channel_id BIGINT NOT NULL COMMENT '频道在Telegram的ID',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '频道标题',
    username VARCHAR(255) DEFAULT '' COMMENT '频道用户名，可选',
    `owner` VARCHAR(255) DEFAULT '' COMMENT '维护人',
    type INT DEFAULT 0 COMMENT '频道类型',
    `status` INT DEFAULT 1 COMMENT '状态：1.启用 2.禁用',
    created_at BIGINT DEFAULT 0 COMMENT '创建时间戳',
    updated_at BIGINT DEFAULT 0 COMMENT '更新时间戳',
    UNIQUE KEY uk_channel_tg_channel_id (tg_channel_id),
    UNIQUE KEY uk_channel_username (username(191)),
    KEY idx_channel_owner (`owner`(191)),
    KEY idx_channel_status (`status`),
    KEY idx_channel_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='telegram频道表';

CREATE TABLE telegram_group (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID，自增',
    tg_group_id BIGINT NOT NULL COMMENT '群组在Telegram的ID',
    title VARCHAR(255) NOT NULL DEFAULT '' COMMENT '群组标题',
    username VARCHAR(255) DEFAULT '' COMMENT '群组用户名，可选',
    `owner` VARCHAR(255) DEFAULT '' COMMENT '维护人',
    type INT DEFAULT 0 COMMENT '群组类型',
    `status` INT DEFAULT 1 COMMENT '状态：1.启用 2.禁用',
    created_at BIGINT DEFAULT 0 COMMENT '创建时间戳',
    updated_at BIGINT DEFAULT 0 COMMENT '更新时间戳',
    UNIQUE KEY uk_group_tg_group_id (tg_group_id),
    UNIQUE KEY uk_group_username (username(191)),
    KEY idx_group_owner (`owner`(191)),
    KEY idx_group_status (`status`),
    KEY idx_group_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='telegram群组表';

CREATE TABLE telegram_group_topic (
    tg_group_id BIGINT NOT NULL COMMENT '所属群组Telegram ID',
    topic_id BIGINT NOT NULL COMMENT '话题ID，对应message_thread_id',
    name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '话题名称',
    created_at BIGINT DEFAULT 0 COMMENT '创建时间戳',
    updated_at BIGINT DEFAULT 0 COMMENT '更新时间戳',
    UNIQUE KEY uk_group_topic_unique (tg_group_id, topic_id),
    KEY idx_group_topic_name (name(191)),
    KEY idx_group_topic_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='telegram群话题表';

CREATE TABLE operate_log (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '日志主键',
    operator VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作者名称',
    operate_at BIGINT NOT NULL COMMENT '操作时间，Unix时间戳',
    ip_address VARCHAR(64) DEFAULT '' COMMENT '操作者IP地址',
    op_type TINYINT NOT NULL DEFAULT 0 COMMENT '操作类型：1-创建，2-更新，3-删除',
    module_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '模块名称',
    target_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '目标ID',
    content TEXT COMMENT '操作内容，JSON格式',
    remark VARCHAR(255) DEFAULT '' COMMENT '备注',
    PRIMARY KEY (id),
    KEY idx_module_target (module_name, target_id),
    KEY idx_operate_at (operate_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='操作日志表';

CREATE TABLE auth (
    username VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
    token VARCHAR(255) NOT NULL DEFAULT '' COMMENT '访问令牌',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用，2-禁用',
    created_at BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    UNIQUE KEY uk_auth_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='接口鉴权表';
