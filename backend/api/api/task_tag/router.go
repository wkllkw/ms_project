package task_tag

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTaskTag struct{}

func init() {
	log.Println("init task_tag router")
	ru := &RouterTaskTag{}
	router.Register(ru)
}

func (*RouterTaskTag) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task_tag", h.list)
	group.POST("/task_tag/save", h.save)
	group.POST("/task_tag/edit", h.edit)
	group.POST("/task_tag/delete", h.delete)
}
