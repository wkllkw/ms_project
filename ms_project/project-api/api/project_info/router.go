package project_info

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectInfo struct {
}

func init() {
	log.Println("init project_info router")
	ru := &RouterProjectInfo{}
	router.Register(ru)
}

func (*RouterProjectInfo) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/project_info", h.list)
	group.POST("/project_info/save", h.save)
	group.POST("/project_info/edit", h.edit)
	group.POST("/project_info/delete", h.del)
}
