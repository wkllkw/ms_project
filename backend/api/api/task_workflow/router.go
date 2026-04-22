package task_workflow

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTaskWorkflow struct {
}

func init() {
	log.Println("init task_workflow router")
	ru := &RouterTaskWorkflow{}
	router.Register(ru)
}

func (*RouterTaskWorkflow) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task_workflow", h.list)
	group.POST("/task_workflow/save", h.save)
	group.POST("/task_workflow/edit", h.edit)
	group.POST("/task_workflow/delete", h.del)
	group.POST("/task_workflow/_getTaskWorkflowRules", h.getTaskWorkflowRules)
}
