package main

import (
	"github.com/gin-gonic/gin"
	_ "test.com/project-api/api"
	"test.com/project-api/config"
	"test.com/project-api/internal/database/migrate"
	"test.com/project-api/router"
	srv "test.com/project-common"
	"net/http"
)

func main() {
	r := gin.Default()
	_ = migrate.AutoMigrate()
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, (&srv.Result{}).Fail(404, "page not found"))
	})
	//路由
	router.InitRouter(r)
	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, nil)
}
