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
	// 节点保存/清理需要 system.node 节点权限
	adminGroup := group.Group("")
	adminGroup.Use(midd.NodeVerify("system.node"))
	adminGroup.POST("/node/allList", h.allList)
	adminGroup.POST("/node/save", h.save)
	adminGroup.POST("/node/clear", h.clear)
}

