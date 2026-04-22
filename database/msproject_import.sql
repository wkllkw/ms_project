-- MySQL dump 10.13  Distrib 8.0.20, for Linux (x86_64)
--
-- Host: localhost    Database: msproject
-- ------------------------------------------------------
-- Server version	8.0.20

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `ms_department`
--

DROP TABLE IF EXISTS `ms_department`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_department` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `parent_id` bigint DEFAULT '0' COMMENT ' id',
  `organization_code` bigint DEFAULT '0' COMMENT ' id',
  `sort` int DEFAULT '0',
  `create_time` bigint DEFAULT '0',
  `deleted` tinyint DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_department`
--

LOCK TABLES `ms_department` WRITE;
/*!40000 ALTER TABLE `ms_department` DISABLE KEYS */;
INSERT INTO `ms_department` (`id`, `name`, `parent_id`, `organization_code`, `sort`, `create_time`, `deleted`) VALUES (1,'寮€鍙戜竴缁?,0,8,0,1773331998685,0),(2,'寮€鍙戜簩缁?,0,8,0,1773332008745,1);
/*!40000 ALTER TABLE `ms_department` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_department_member`
--

DROP TABLE IF EXISTS `ms_department_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_department_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `department_id` bigint DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_department_member_department_id` (`department_id`),
  KEY `idx_ms_department_member_member_id` (`member_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_department_member`
--

LOCK TABLES `ms_department_member` WRITE;
/*!40000 ALTER TABLE `ms_department_member` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_department_member` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_invite_link`
--

DROP TABLE IF EXISTS `ms_invite_link`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_invite_link` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL,
  `invite_code` varchar(64) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `expired_at` bigint DEFAULT NULL,
  `create_by` bigint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ms_invite_link_invite_code` (`invite_code`),
  KEY `idx_ms_invite_link_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_invite_link`
--

LOCK TABLES `ms_invite_link` WRITE;
/*!40000 ALTER TABLE `ms_invite_link` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_invite_link` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_member`
--

DROP TABLE IF EXISTS `ms_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `account` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `password` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT '',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT '',
  `mobile` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `realname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `status` tinyint(1) DEFAULT '0',
  `last_login_time` bigint DEFAULT NULL,
  `sex` tinyint DEFAULT '0',
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT '',
  `idcard` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `province` int DEFAULT '0',
  `city` int DEFAULT '0',
  `area` int DEFAULT '0',
  `address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `dingtalk_openid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT ' openid',
  `dingtalk_unionid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT ' unionid',
  `dingtalk_userid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT ' id',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1004 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_member`
--

LOCK TABLES `ms_member` WRITE;
/*!40000 ALTER TABLE `ms_member` DISABLE KEYS */;
INSERT INTO `ms_member` (`id`, `account`, `password`, `name`, `mobile`, `realname`, `create_time`, `status`, `last_login_time`, `sex`, `avatar`, `idcard`, `province`, `city`, `area`, `address`, `description`, `email`, `dingtalk_openid`, `dingtalk_unionid`, `dingtalk_userid`) VALUES (1000,'admin','e10adc3949ba59abbe56e057f20f883e','',NULL,NULL,1773165136,1,NULL,0,'',NULL,0,0,0,NULL,NULL,'admin@example.com',NULL,NULL,NULL),(1003,'rookie','91b87eef27f5abdc35dc6295e369ba2b','rookie','18089176282','',1773166025395,1,1773166025395,0,'','',0,0,0,'','','wankailei@qq.com','','','');
/*!40000 ALTER TABLE `ms_member` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_notify`
--

DROP TABLE IF EXISTS `ms_notify`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_notify` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `member_code` bigint DEFAULT NULL,
  `title` longtext COLLATE utf8mb4_general_ci,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `type` tinyint DEFAULT NULL,
  `is_read` tinyint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `action` longtext COLLATE utf8mb4_general_ci,
  `send_data` text COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_ms_notify_member_code` (`member_code`),
  KEY `idx_ms_notify_type` (`type`),
  KEY `idx_ms_notify_is_read` (`is_read`),
  KEY `idx_ms_notify_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_notify`
--

LOCK TABLES `ms_notify` WRITE;
/*!40000 ALTER TABLE `ms_notify` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_notify` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_organization`
--

DROP TABLE IF EXISTS `ms_organization`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_organization` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `avatar` varchar(511) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `personal` tinyint(1) DEFAULT '0',
  `address` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `province` int DEFAULT '0',
  `city` int DEFAULT '0',
  `area` int DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_organization`
--

LOCK TABLES `ms_organization` WRITE;
/*!40000 ALTER TABLE `ms_organization` DISABLE KEYS */;
INSERT INTO `ms_organization` (`id`, `name`, `avatar`, `description`, `member_id`, `create_time`, `personal`, `address`, `province`, `city`, `area`) VALUES (8,'rookie涓汉缁勭粐','https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5','',1003,1773166025398,1,'',0,0,0);
/*!40000 ALTER TABLE `ms_organization` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project`
--

DROP TABLE IF EXISTS `ms_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `cover` varchar(511) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` varchar(511) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `access_control_type` int DEFAULT NULL COMMENT ' 1  2 ',
  `white_list` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `sort` int DEFAULT NULL,
  `deleted` int DEFAULT NULL COMMENT ' 0  1 ',
  `template_code` int DEFAULT NULL COMMENT ' code',
  `schedule` double DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `organization_code` bigint DEFAULT NULL,
  `deleted_time` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `private` int DEFAULT NULL COMMENT ' 0  1 ',
  `prefix` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `open_prefix` int DEFAULT NULL,
  `archive` int DEFAULT NULL COMMENT ' 0  1 ',
  `archive_time` bigint DEFAULT NULL,
  `open_begin_time` int DEFAULT NULL,
  `open_task_private` int DEFAULT NULL,
  `task_board_theme` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `begin_time` bigint DEFAULT NULL,
  `end_time` bigint DEFAULT NULL,
  `auto_update_schedule` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_org_code` (`organization_code`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project`
--

LOCK TABLES `ms_project` WRITE;
/*!40000 ALTER TABLE `ms_project` DISABLE KEYS */;
INSERT INTO `ms_project` (`id`, `cover`, `name`, `description`, `access_control_type`, `white_list`, `sort`, `deleted`, `template_code`, `schedule`, `create_time`, `organization_code`, `deleted_time`, `private`, `prefix`, `open_prefix`, `archive`, `archive_time`, `open_begin_time`, `open_task_private`, `task_board_theme`, `begin_time`, `end_time`, `auto_update_schedule`) VALUES (1,'https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500','杩愮淮','棰濆浼侀箙鍘?,0,'',0,1,0,0,1773334414463,8,'',0,'',0,0,0,0,0,'simple',0,0,0),(2,'https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500','sre','澶ф拻',0,'',0,0,0,0,1773334579510,8,'',0,'',0,1,1773340100261,0,0,'simple',0,0,0);
/*!40000 ALTER TABLE `ms_project` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_auth`
--

DROP TABLE IF EXISTS `ms_project_auth`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_auth` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` longtext COLLATE utf8mb4_general_ci,
  `desc` longtext COLLATE utf8mb4_general_ci,
  `status` bigint DEFAULT NULL,
  `is_default` bigint DEFAULT NULL,
  `create_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_project_auth_create_at` (`create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_auth`
--

LOCK TABLES `ms_project_auth` WRITE;
/*!40000 ALTER TABLE `ms_project_auth` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_project_auth` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_auth_node`
--

DROP TABLE IF EXISTS `ms_project_auth_node`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_auth_node` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `auth_id` bigint DEFAULT NULL,
  `node` longtext COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`),
  KEY `idx_ms_project_auth_node_auth_id` (`auth_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_auth_node`
--

LOCK TABLES `ms_project_auth_node` WRITE;
/*!40000 ALTER TABLE `ms_project_auth_node` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_project_auth_node` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_collection`
--

DROP TABLE IF EXISTS `ms_project_collection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_collection` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint NOT NULL COMMENT ' ID',
  `member_code` bigint NOT NULL COMMENT ' ID',
  `create_time` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_member_code` (`member_code`),
  KEY `idx_project_code` (`project_code`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_collection`
--

LOCK TABLES `ms_project_collection` WRITE;
/*!40000 ALTER TABLE `ms_project_collection` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_project_collection` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_event`
--

DROP TABLE IF EXISTS `ms_project_event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_event` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT '0' COMMENT ' id',
  `member_code` bigint DEFAULT '0' COMMENT ' id',
  `event_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `event_content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `create_time` bigint DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_event`
--

LOCK TABLES `ms_project_event` WRITE;
/*!40000 ALTER TABLE `ms_project_event` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_project_event` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_events`
--

DROP TABLE IF EXISTS `ms_project_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL,
  `title` longtext COLLATE utf8mb4_general_ci,
  `description` text COLLATE utf8mb4_general_ci,
  `begin_time` longtext COLLATE utf8mb4_general_ci,
  `end_time` longtext COLLATE utf8mb4_general_ci,
  `all_day` tinyint DEFAULT NULL,
  `position` longtext COLLATE utf8mb4_general_ci,
  `create_by` bigint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `deleted` tinyint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_project_events_project_code` (`project_code`),
  KEY `idx_ms_project_events_deleted` (`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_events`
--

LOCK TABLES `ms_project_events` WRITE;
/*!40000 ALTER TABLE `ms_project_events` DISABLE KEYS */;
INSERT INTO `ms_project_events` (`id`, `project_code`, `title`, `description`, `begin_time`, `end_time`, `all_day`, `position`, `create_by`, `create_time`, `deleted`) VALUES (1,2,'鐨勬拻','鐨勬拻鏃?,'2026-03-13 02:47:17','2026-03-13 02:47:17',0,'',1003,1773341244712,1),(2,2,'32131311','鑰屼笖閮芥槸椤惰捣鎾?,'2026-03-06 02:47:42','2026-03-06 02:47:42',0,'鑰屾垜鍗?,1003,1773341273514,1),(3,2,'鍑虹幇鍦?,'鐨勬拻','2026-03-13 02:58:17','2026-03-13 02:58:17',0,'',1003,1773341909506,1),(4,2,'鎵撴拻','','2026-03-06 04:53:21','2026-03-06 04:53:21',0,'',1003,1773348809773,1);
/*!40000 ALTER TABLE `ms_project_events` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_events_member`
--

DROP TABLE IF EXISTS `ms_project_events_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_events_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `events_id` bigint DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_project_events_member_events_id` (`events_id`),
  KEY `idx_ms_project_events_member_member_id` (`member_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_events_member`
--

LOCK TABLES `ms_project_events_member` WRITE;
/*!40000 ALTER TABLE `ms_project_events_member` DISABLE KEYS */;
INSERT INTO `ms_project_events_member` (`id`, `events_id`, `member_id`, `status`) VALUES (1,1,1003,1),(2,2,1003,1),(3,3,1003,1),(4,1,1000,1),(5,4,1003,1);
/*!40000 ALTER TABLE `ms_project_events_member` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_member`
--

DROP TABLE IF EXISTS `ms_project_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL,
  `member_code` bigint DEFAULT NULL,
  `join_time` bigint DEFAULT NULL,
  `is_owner` bigint DEFAULT NULL,
  `authorize` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_project_code` (`project_code`),
  KEY `idx_member_code` (`member_code`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_member`
--

LOCK TABLES `ms_project_member` WRITE;
/*!40000 ALTER TABLE `ms_project_member` DISABLE KEYS */;
INSERT INTO `ms_project_member` (`id`, `project_code`, `member_code`, `join_time`, `is_owner`, `authorize`) VALUES (1,1,1003,1773334414468,1,''),(2,2,1003,1773334579513,1,'');
/*!40000 ALTER TABLE `ms_project_member` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_menu`
--

DROP TABLE IF EXISTS `ms_project_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pid` bigint DEFAULT NULL,
  `title` longtext,
  `icon` longtext,
  `url` longtext,
  `file_path` longtext,
  `params` longtext,
  `node` longtext,
  `sort` bigint DEFAULT NULL,
  `status` bigint DEFAULT NULL,
  `create_by` bigint DEFAULT NULL,
  `is_inner` bigint DEFAULT NULL,
  `values` longtext,
  `show_slider` bigint DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_ms_project_menu_pid` (`pid`)
) ENGINE=InnoDB AUTO_INCREMENT=176 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_menu`
--

LOCK TABLES `ms_project_menu` WRITE;
/*!40000 ALTER TABLE `ms_project_menu` DISABLE KEYS */;
INSERT INTO `ms_project_menu` (`id`, `pid`, `title`, `icon`, `url`, `file_path`, `params`, `node`, `sort`, `status`, `create_by`, `is_inner`, `values`, `show_slider`) VALUES (1,0,'宸ヤ綔鍙?,'home','home','home/index','','home',1,1,0,0,'',1),(10,0,'椤圭洰','project','project/list','project/list/index','my','project',2,1,0,0,'',1),(11,10,'椤圭洰鍒楄〃','project','project/list','project/list/index','my','project.list',1,1,0,0,'',1),(12,10,'椤圭洰妯℃澘','appstore','project/template','project/template/index','','project.template',2,1,0,0,'',1),(20,0,'閫氱煡','bell','notify/notice','notify/notice','','notify',2,1,0,0,'',1),(21,20,'閫氱煡鍒楄〃','bell','notify/notice','notify/notice','','notify.notice',1,1,0,0,'',1),(22,20,'绯荤粺娑堟伅','notification','notify/system','notify/system','','notify.system',2,1,0,0,'',1),(30,0,'鍥㈤槦','team','members','members/index','','members',3,1,0,0,'',1),(31,30,'鎴愬憳','team','members','members/index','','members.index',1,1,0,0,'',1),(40,0,'绯荤粺','setting','system/account','system/account/index','','system',4,1,0,0,'',1),(41,40,'璐﹀彿绠＄悊','user','system/account','system/account/index','','system.account',1,1,0,0,'',1),(42,40,'鑿滃崟绠＄悊','menu','system/config/menu','system/config/menu','','system.menu',2,1,0,0,'',1),(43,40,'鑺傜偣绠＄悊','apartment','system/config/node','system/config/node','','system.node',3,1,0,0,'',1);
/*!40000 ALTER TABLE `ms_project_menu` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_project_template`
--

DROP TABLE IF EXISTS `ms_project_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_project_template` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` varchar(511) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `organization_code` bigint DEFAULT NULL,
  `cover` varchar(511) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `member_code` bigint DEFAULT NULL,
  `is_system` int DEFAULT NULL COMMENT ' 0  1 ',
  PRIMARY KEY (`id`),
  KEY `idx_org_code` (`organization_code`),
  KEY `idx_is_system` (`is_system`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_project_template`
--

LOCK TABLES `ms_project_template` WRITE;
/*!40000 ALTER TABLE `ms_project_template` DISABLE KEYS */;
INSERT INTO `ms_project_template` (`id`, `name`, `description`, `sort`, `create_time`, `organization_code`, `cover`, `member_code`, `is_system`) VALUES (1,'产品进展','适用于互联网产品人员对产品计划、跟进及发布管理',1,1773168102000,NULL,'https://img2.baidu.com/it/u=2241642503,1613686234&fm=253&fmt=auto&app=138&f=JPEG?w=603&h=500',NULL,1),(2,'需求管理','适用于产品部门对需求的收集、评估及反馈管理',2,1773168102000,NULL,'https://img0.baidu.com/it/u=437485064,4277010738&fm=253&fmt=auto&app=138&f=JPEG?w=610&h=491',NULL,1),(3,'机械制造','适用于制造商对图纸设计及制造安装的工作流程管理',3,1773168102000,NULL,'https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fbpic.51yuansu.com%2Fpic2%2Fcover%2F00%2F38%2F93%2F5812ca7a24020_610.jpg',NULL,1);
/*!40000 ALTER TABLE `ms_project_template` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task`
--

DROP TABLE IF EXISTS `ms_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL,
  `name` longtext COLLATE utf8mb4_general_ci,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `status` tinyint DEFAULT NULL,
  `priority` tinyint DEFAULT NULL,
  `begin_time` bigint DEFAULT NULL,
  `end_time` bigint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `member_code` bigint DEFAULT NULL,
  `owner_code` bigint DEFAULT NULL,
  `stage_code` bigint DEFAULT NULL,
  `parent_task_id` bigint DEFAULT NULL,
  `sort` bigint DEFAULT NULL,
  `deleted` tinyint DEFAULT NULL,
  `private` tinyint DEFAULT NULL,
  `assign_to` bigint DEFAULT NULL,
  `done` tinyint DEFAULT NULL,
  `like` bigint DEFAULT NULL,
  `star` bigint DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task`
--

LOCK TABLES `ms_task` WRITE;
/*!40000 ALTER TABLE `ms_task` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_comment`
--

DROP TABLE IF EXISTS `ms_task_comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_comment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  `comment` text COLLATE utf8mb4_general_ci,
  `create_time` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_task_comment_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_comment`
--

LOCK TABLES `ms_task_comment` WRITE;
/*!40000 ALTER TABLE `ms_task_comment` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task_comment` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_member`
--

DROP TABLE IF EXISTS `ms_task_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_task_member_task_id` (`task_id`),
  KEY `idx_ms_task_member_member_id` (`member_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_member`
--

LOCK TABLES `ms_task_member` WRITE;
/*!40000 ALTER TABLE `ms_task_member` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task_member` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_stages`
--

DROP TABLE IF EXISTS `ms_task_stages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_stages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL,
  `name` longtext COLLATE utf8mb4_general_ci,
  `sort` bigint DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `deleted` tinyint DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_stages`
--

LOCK TABLES `ms_task_stages` WRITE;
/*!40000 ALTER TABLE `ms_task_stages` DISABLE KEYS */;
INSERT INTO `ms_task_stages` (`id`, `project_code`, `name`, `sort`, `create_time`, `deleted`) VALUES (1,1,'涓夊ぇ',1,1773348770682,0);
/*!40000 ALTER TABLE `ms_task_stages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_stages_template`
--

DROP TABLE IF EXISTS `ms_task_stages_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_stages_template` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `project_template_code` int DEFAULT NULL,
  `create_time` bigint DEFAULT NULL,
  `sort` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_template_code` (`project_template_code`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_stages_template`
--

LOCK TABLES `ms_task_stages_template` WRITE;
/*!40000 ALTER TABLE `ms_task_stages_template` DISABLE KEYS */;
INSERT INTO `ms_task_stages_template` (`id`, `name`, `project_template_code`, `create_time`, `sort`) VALUES (1,'产品计划',1,1773168108000,1),(2,'即将发布',1,1773168108000,2),(3,'测试',1,1773168108000,3),(4,'准备发布',1,1773168108000,4),(5,'发布成功',1,1773168108000,5),(6,'需求收集',2,1773168108000,1),(7,'评估确认',2,1773168108000,2),(8,'需求暂缓',2,1773168108000,3),(9,'研发中',2,1773168108000,4),(10,'内测中',2,1773168108000,5),(11,'通知用户',2,1773168108000,6),(12,'已完成&归档',2,1773168108000,7),(13,'协议签订',3,1773168108000,1),(14,'图纸设计',3,1773168108000,2),(15,'评审及打样',3,1773168108000,3),(16,'构件采购',3,1773168108000,4),(17,'制造安装',3,1773168108000,5),(18,'内部检验',3,1773168108000,6),(19,'验收',3,1773168108000,7);
/*!40000 ALTER TABLE `ms_task_stages_template` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_tag`
--

DROP TABLE IF EXISTS `ms_task_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_code` bigint DEFAULT NULL,
  `name` longtext COLLATE utf8mb4_general_ci,
  `color` longtext COLLATE utf8mb4_general_ci,
  `create_time` bigint DEFAULT NULL,
  `deleted` tinyint DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_tag`
--

LOCK TABLES `ms_task_tag` WRITE;
/*!40000 ALTER TABLE `ms_task_tag` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task_tag` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_tag_rel`
--

DROP TABLE IF EXISTS `ms_task_tag_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_tag_rel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint DEFAULT NULL,
  `tag_id` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_task_tag_rel_task_id` (`task_id`),
  KEY `idx_ms_task_tag_rel_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_tag_rel`
--

LOCK TABLES `ms_task_tag_rel` WRITE;
/*!40000 ALTER TABLE `ms_task_tag_rel` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task_tag_rel` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ms_task_work_time`
--

DROP TABLE IF EXISTS `ms_task_work_time`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ms_task_work_time` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint DEFAULT NULL,
  `member_id` bigint DEFAULT NULL,
  `work_time` bigint DEFAULT NULL,
  `remark` longtext COLLATE utf8mb4_general_ci,
  `create_time` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ms_task_work_time_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ms_task_work_time`
--

LOCK TABLES `ms_task_work_time` WRITE;
/*!40000 ALTER TABLE `ms_task_work_time` DISABLE KEYS */;
/*!40000 ALTER TABLE `ms_task_work_time` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'msproject'
--

--
-- Dumping routines for database 'msproject'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-03-13 19:33:24
