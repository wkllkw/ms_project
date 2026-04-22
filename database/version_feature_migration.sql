-- 版本管理功能增强迁移脚本
-- 添加 version_code 字段到 ms_task 表，创建 ms_project_version_log 表

-- 1. 为 ms_task 表添加 version_code 字段
ALTER TABLE `ms_task` ADD COLUMN `version_code` bigint DEFAULT 0 COMMENT '版本ID' AFTER `parent_task_id`;

-- 2. 为 ms_task 表添加 features_code 字段
ALTER TABLE `ms_task` ADD COLUMN `features_code` bigint DEFAULT 0 COMMENT '版本库ID' AFTER `version_code`;

-- 3. 创建版本日志表
CREATE TABLE IF NOT EXISTS `ms_project_version_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `member_code` bigint DEFAULT NULL COMMENT '操作人ID',
  `source_code` bigint DEFAULT NULL COMMENT '版本ID',
  `content` text COMMENT '操作内容',
  `remark` text COMMENT '日志描述',
  `type` varchar(50) DEFAULT NULL COMMENT '操作类型',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  `icon` varchar(50) DEFAULT NULL COMMENT '图标',
  `features_code` bigint DEFAULT NULL COMMENT '版本库ID',
  PRIMARY KEY (`id`),
  KEY `idx_version_log_source_code` (`source_code`),
  KEY `idx_version_log_features_code` (`features_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='版本操作日志表';

-- 4. 为 ms_project_version 表添加索引（如果不存在）
-- ALTER TABLE `ms_project_version` ADD INDEX `idx_version_features_code` (`features_code`);
