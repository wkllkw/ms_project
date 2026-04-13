-- 创建缺失的数据库表

-- 1. 文件管理表
CREATE TABLE IF NOT EXISTS `ms_file` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL COMMENT '项目ID',
  `member_code` bigint DEFAULT NULL COMMENT '上传者ID',
  `title` varchar(255) DEFAULT NULL COMMENT '文件标题',
  `file_name` varchar(255) DEFAULT NULL COMMENT '原始文件名',
  `file_type` varchar(50) DEFAULT NULL COMMENT '文件类型',
  `file_size` bigint DEFAULT 0 COMMENT '文件大小(字节)',
  `file_url` varchar(500) DEFAULT NULL COMMENT '文件URL',
  `file_path` varchar(500) DEFAULT NULL COMMENT '文件存储路径',
  `description` text COMMENT '文件描述',
  `deleted` tinyint DEFAULT 0 COMMENT '是否删除 0否 1是',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  `update_time` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_file_project_code` (`project_code`),
  KEY `idx_file_member_code` (`member_code`),
  KEY `idx_file_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目文件表';

-- 2. 项目版本库表 (project_features)
CREATE TABLE IF NOT EXISTS `ms_project_features` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL COMMENT '项目ID',
  `name` varchar(255) DEFAULT NULL COMMENT '版本库名称',
  `description` text COMMENT '描述',
  `sort` int DEFAULT 0 COMMENT '排序',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  `update_time` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_features_project_code` (`project_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目版本库表';

-- 3. 项目版本表 (project_version)
CREATE TABLE IF NOT EXISTS `ms_project_version` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `features_code` bigint DEFAULT NULL COMMENT '版本库ID',
  `project_code` bigint DEFAULT NULL COMMENT '项目ID',
  `name` varchar(255) DEFAULT NULL COMMENT '版本名称',
  `description` text COMMENT '版本描述',
  `start_time` bigint DEFAULT NULL COMMENT '开始时间',
  `plan_publish_time` bigint DEFAULT NULL COMMENT '计划发布时间',
  `publish_time` bigint DEFAULT NULL COMMENT '实际发布时间',
  `status` tinyint DEFAULT 0 COMMENT '状态 0未发布 1已发布',
  `sort` int DEFAULT 0 COMMENT '排序',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  `update_time` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_version_features_code` (`features_code`),
  KEY `idx_version_project_code` (`project_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目版本表';

-- 4. 任务工作流表 (task_workflow)
CREATE TABLE IF NOT EXISTS `ms_task_workflow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL COMMENT '项目ID',
  `name` varchar(255) DEFAULT NULL COMMENT '工作流名称',
  `description` text COMMENT '描述',
  `rules` text COMMENT '规则JSON',
  `sort` int DEFAULT 0 COMMENT '排序',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  `update_time` bigint DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_workflow_project_code` (`project_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='任务工作流表';

-- 5. 资源链接表 (source_link)
CREATE TABLE IF NOT EXISTS `ms_source_link` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_code` bigint DEFAULT NULL COMMENT '任务ID',
  `member_code` bigint DEFAULT NULL COMMENT '创建者ID',
  `title` varchar(255) DEFAULT NULL COMMENT '链接标题',
  `url` varchar(500) DEFAULT NULL COMMENT '链接URL',
  `description` text COMMENT '描述',
  `sort` int DEFAULT 0 COMMENT '排序',
  `create_time` bigint DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_source_task_code` (`task_code`),
  KEY `idx_source_member_code` (`member_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='任务资源链接表';
