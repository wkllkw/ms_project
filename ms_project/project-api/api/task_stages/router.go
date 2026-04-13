package task_stages

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTaskStages struct {
}

func init() {
	log.Println("init task_stages router")
	ru := &RouterTaskStages{}
	router.Register(ru)
}

func (*RouterTaskStages) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task_stages", h.list)
	group.POST("/task_stages/taskTree", h.taskTree)
	group.POST("/task_stages/_getAll", h.getAll)
	group.POST("/task_stages/tasks", h.tasks)
	group.POST("/task_stages/sort", h.sort)
	group.POST("/task_stages/save", h.save)
	group.POST("/task_stages/edit", h.edit)
	group.POST("/task_stages/delete", h.del)
}

