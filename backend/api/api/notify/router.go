package notify

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterNotify struct {
}

func init() {
	log.Println("init notify router")
	ru := &RouterNotify{}
	router.Register(ru)
}

func (*RouterNotify) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/notify", h.list)
	group.POST("/notify/noReads", h.noReads)
	group.POST("/notify/_clearAll", h.clearAll)
	group.POST("/notify/save", h.save)
	group.POST("/notify/edit", h.edit)
	group.POST("/notify/delete", h.del)
	group.POST("/notify/batchDel", h.batchDel)
	group.POST("/notify/setReadied", h.setReadied)
}

