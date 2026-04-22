package project_features

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectFeatures struct {
}

func init() {
	log.Println("init project_features router")
	ru := &RouterProjectFeatures{}
	router.Register(ru)
}

func (*RouterProjectFeatures) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/project_features", h.list)
	group.POST("/project_features/save", h.save)
	group.POST("/project_features/edit", h.edit)
	group.POST("/project_features/delete", h.del)
}
