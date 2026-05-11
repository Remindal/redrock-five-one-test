CREATE DATABASE IF NOT EXISTS `seckill` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  USE `seckill`;

CREATE TABLE IF NOT EXISTS `activity` (
                                          `id` BIGINT NOT NULL AUTO_INCREMENT,
                                          `name` VARCHAR(255) NOT NULL,
    `stock` INT NOT NULL DEFAULT 0,
    `remain_stock` INT NOT NULL DEFAULT 0,
    `start_time` DATETIME NOT NULL,
    `end_time` DATETIME NOT NULL,
    `status` INT NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `order` (
                                       `id` BIGINT NOT NULL AUTO_INCREMENT,
                                       `activity_id` BIGINT NOT NULL,
                                       `user_id` VARCHAR(64) NOT NULL,
    `status` VARCHAR(32) NOT NULL DEFAULT 'PROCESSING',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_activity_user` (`activity_id`, `user_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;