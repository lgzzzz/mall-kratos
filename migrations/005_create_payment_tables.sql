-- 005_create_payment_tables.sql
CREATE DATABASE IF NOT EXISTS mall_payment DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE mall_payment;

CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    status INT NOT NULL DEFAULT 1,
    transaction_id VARCHAR(100),
    payment_url VARCHAR(500),
    expires_at DATETIME,
    paid_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_order_id (order_id),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
