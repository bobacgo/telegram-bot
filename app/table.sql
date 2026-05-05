CREATE TABLE telegram_bots (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID，自增',
    bot_tg_id BIGINT NOT NULL COMMENT '机器人在Telegram的ID',
    username VARCHAR(255) DEFAULT '' COMMENT '机器人用户名',
    token VARCHAR(255) NOT NULL COMMENT '机器人Token',
    webhook_secret VARCHAR(255) DEFAULT '' COMMENT 'Webhook认证密钥，对应Telegram的X-Telegram-Bot-Api-Secret-Token头',
    `owner` VARCHAR(255) DEFAULT '' COMMENT '机器人拥有者的Telegram用户名',
    type INT DEFAULT 0 COMMENT '机器人类型',
    `status` INT DEFAULT 1 COMMENT '状态：1.启用 2.禁用 3.封禁',
    created_at BIGINT DEFAULT 0 COMMENT '创建时间戳',
    updated_at BIGINT DEFAULT 0 COMMENT '更新时间戳',
    
    -- 唯一索引约束
    UNIQUE KEY uk_bot_tg_id (bot_tg_id),
    UNIQUE KEY uk_username (username(191)),
    
    -- 普通索引
    KEY idx_owner (owner(191)),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='telegram机器人表';