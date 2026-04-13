package task_stages_template

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTaskStagesTemplate struct {
}

func init() {
	log.Println("init task_stages_template router")
	ru := &RouterTaskStagesTemplate{}
	router.Register(ru)
}

func (*RouterTaskStagesTemplate) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task_stages_template", h.list)
	group.POST("/task_stages_template/save", h.save)
	group.POST("/task_stages_template/edit", h.edit)
	group.POST("/task_stages_template/delete", h.del)
}

