package project_collect

import (
	"log"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectCollect struct{}

func init() {
	log.Println("init project_collect router")
	ru := &RouterProjectCollect{}
	router.Register(ru)
}

func (*RouterProjectCollect) Route(r *gin.Engine) {
	h := New()
	group := r.Group("/project/project_collect")
	group.Use(midd.TokenVerify())
	group.POST("/collect", h.collectProject)
	group.POST("/list", h.getCollectionList)
}
