package source_link

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterSourceLink struct {
}

func init() {
	log.Println("init source_link router")
	ru := &RouterSourceLink{}
	router.Register(ru)
}

func (*RouterSourceLink) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/source_link", h.list)
	group.POST("/source_link/save", h.save)
	group.POST("/source_link/edit", h.edit)
	group.POST("/source_link/delete", h.del)
}
