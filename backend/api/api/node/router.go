package node

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterNode struct {
}

func init() {
	log.Println("init node router")
	ru := &RouterNode{}
	router.Register(ru)
}

func (*RouterNode) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/node", h.list)
	group.POST("/node/allList", h.allList)
	group.POST("/node/save", h.save)
	group.POST("/node/clear", h.clear)
}

