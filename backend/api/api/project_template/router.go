package project_template

import (
	"log"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectTemplate struct{}

func init() {
	log.Println("init project_template router")
	ru := &RouterProjectTemplate{}
	router.Register(ru)
}

func (*RouterProjectTemplate) Route(r *gin.Engine) {
	h := New()
	group := r.Group("/project/project_template")
	group.Use(midd.TokenVerify())
	group.POST("", h.getTemplateList)
	group.POST("/save", h.createTemplate)
	group.POST("/edit", h.editTemplate)
	group.POST("/delete", h.deleteTemplate)
	group.POST("/uploadCover", h.uploadCover)
}
