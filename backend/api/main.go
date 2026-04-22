package main

import (
	"github.com/gin-gonic/gin"
	_ "test.com/project-api/api" // 导入 api 包触发路由注册
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "test.com/project-api/docs" // swagger 生成的文档
	"test.com/project-api/config"
	"test.com/project-api/internal/cache"
	"test.com/project-api/internal/database/migrate"
	"test.com/project-api/internal/email"
	"test.com/project-api/internal/scheduler"
	"test.com/project-api/router"
	srv "test.com/project-common"
	"log"
	"net/http"
)

// @title MS Project 任务协作平台 API
// @version 1.0
// @description 基于 Go + Gin + Vue 的企业级任务协作系统后端API<br/>
// @description 功能模块: 项目管理 | 任务看板 | 甘特图 | 日程安排 | 文件管理 | 成员权限 | 实时通知
// @termsOfService http://swagger.io/terms/

// @contact.name MS Project API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:18000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer Token 认证（登录后获取）

// @tag.name project
// @tag.description 项目管理相关接口

// @tag.name task
// @tag.description 任务管理相关接口（CRUD/分配/评论/工时/附件）

// @tag.name file
// @tag.description 文件上传与管理接口

// @tag.name user
// @tag.description 用户认证与个人设置接口

// @tag.name notify
// @tag.description 系统通知接口

func main() {
	r := gin.Default()
	_ = migrate.AutoMigrate()

	// 静态文件服务：提供 /uploads 路径的文件访问
	r.Static("/uploads", "./uploads")

	// 初始化 Redis 缓存（如果配置了的话）
	if config.C.RedisConfig != nil && config.C.RedisConfig.Addr != "" {
		if err := cache.InitRedis(
			config.C.RedisConfig.Addr,
			config.C.RedisConfig.Password,
			config.C.RedisConfig.Db,
		); err != nil {
			log.Printf("[WARN] redis init failed: %v (cache disabled)", err)
		}
	}

	// 初始化邮件服务（如果配置了的话）
	if config.C.MailConfig != nil && config.C.MailConfig.Host != "" {
		email.Init(&email.Config{
			Host:     config.C.MailConfig.Host,
			Port:     config.C.MailConfig.Port,
			User:     config.C.MailConfig.User,
			Password: config.C.MailConfig.Password,
			From:     config.C.MailConfig.From,
		})
		log.Println("[INFO] email service initialized")
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, (&srv.Result{}).Fail(404, "page not found"))
	})

	//路由
	router.InitRouter(r)

	// Swagger API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动定时任务（后台 goroutine）
	go scheduler.StartTaskReminder()

	log.Println("[INFO] task reminder scheduler started")
	log.Println("[INFO] swagger docs available at http://localhost:18000/swagger/index.html")

	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, nil)
}
