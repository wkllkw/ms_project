SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 修复 ms_project_template 表数据
DELETE FROM `ms_project_template` WHERE `is_system` = 1;
INSERT INTO `ms_project_template` (`id`, `name`, `description`, `sort`, `create_time`, `organization_code`, `cover`, `member_code`, `is_system`) VALUES
(1,'产品进展','适用于互联网产品人员对产品计划、跟进及发布管理',1,1773168102000,NULL,'https://img2.baidu.com/it/u=2241642503,1613686234&fm=253&fmt=auto&app=138&f=JPEG?w=603&h=500',NULL,1),
(2,'需求管理','适用于产品部门对需求的收集、评估及反馈管理',2,1773168102000,NULL,'https://img0.baidu.com/it/u=437485064,4277010738&fm=253&fmt=auto&app=138&f=JPEG?w=610&h=491',NULL,1),
(3,'机械制造','适用于制造商对图纸设计及制造安装的工作流程管理',3,1773168102000,NULL,'https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fbpic.51yuansu.com%2Fpic2%2Fcover%2F00%2F38%2F93%2F5812ca7a24020_610.jpg',NULL,1);

-- 修复 ms_task_stages_template 表数据
DELETE FROM `ms_task_stages_template` WHERE `project_template_code` IN (1, 2, 3);
INSERT INTO `ms_task_stages_template` (`id`, `name`, `project_template_code`, `create_time`, `sort`) VALUES
(1,'产品计划',1,1773168108000,1),
(2,'即将发布',1,1773168108000,2),
(3,'测试',1,1773168108000,3),
(4,'准备发布',1,1773168108000,4),
(5,'发布成功',1,1773168108000,5),
(6,'需求收集',2,1773168108000,1),
(7,'评估确认',2,1773168108000,2),
(8,'需求暂缓',2,1773168108000,3),
(9,'研发中',2,1773168108000,4),
(10,'内测中',2,1773168108000,5),
(11,'通知用户',2,1773168108000,6),
(12,'已完成&归档',2,1773168108000,7),
(13,'协议签订',3,1773168108000,1),
(14,'图纸设计',3,1773168108000,2),
(15,'评审及打样',3,1773168108000,3),
(16,'构件采购',3,1773168108000,4),
(17,'制造安装',3,1773168108000,5),
(18,'内部检验',3,1773168108000,6),
(19,'验收',3,1773168108000,7);
