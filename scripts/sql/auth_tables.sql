-- scripts/sql/auth_tables.sql

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id             BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '雪花ID',
    username       VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash  VARCHAR(255) DEFAULT NULL COMMENT '密码哈希',
    email          VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    email_verified TINYINT(1) DEFAULT 0 COMMENT '邮箱是否验证',
    area_code      VARCHAR(10) DEFAULT NULL COMMENT '区号',
    phone          VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    phone_verified TINYINT(1) DEFAULT 0 COMMENT '手机号是否验证',
    avatar         VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
    nickname       VARCHAR(50) DEFAULT NULL COMMENT '昵称',
    status         TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2停用',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_username (username),
    INDEX idx_email (email),
    UNIQUE INDEX idx_area_phone (area_code, phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 第三方绑定表
CREATE TABLE IF NOT EXISTS third_party_binds (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    provider      TINYINT NOT NULL COMMENT '提供商：1微信 2GitHub',
    provider_id   VARCHAR(100) NOT NULL COMMENT '第三方用户ID',
    provider_name VARCHAR(100) DEFAULT NULL COMMENT '第三方用户名',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_provider_user (provider, provider_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='第三方账号绑定表';
