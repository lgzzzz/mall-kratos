-- 006_create_inventory_tables.sql
CREATE DATABASE IF NOT EXISTS mall_inventory DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE mall_inventory;

CREATE TABLE IF NOT EXISTS inventories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku_id BIGINT NOT NULL,
    warehouse_id BIGINT NOT NULL DEFAULT 1,
    stock BIGINT NOT NULL DEFAULT 0,
    locked_stock BIGINT NOT NULL DEFAULT 0,
    available_stock BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_sku_warehouse (sku_id, warehouse_id),
    INDEX idx_sku_id (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    sku_id BIGINT NOT NULL,
    warehouse_id BIGINT NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    quantity BIGINT NOT NULL,
    before_stock BIGINT NOT NULL,
    after_stock BIGINT NOT NULL,
    reference_id VARCHAR(100),
    remark TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sku_id (sku_id),
    INDEX idx_reference_id (reference_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
