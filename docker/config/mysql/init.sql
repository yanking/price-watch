-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS watch CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 授予 root 用户远程访问权限（开发环境）
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;

-- 使用数据库
USE watch;

-- 示例：创建基础表结构（根据实际需求调整）
-- CREATE TABLE IF NOT EXISTS prices (
--     id BIGINT PRIMARY KEY AUTO_INCREMENT,
--     symbol VARCHAR(20) NOT NULL,
--     price DECIMAL(20, 8) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
