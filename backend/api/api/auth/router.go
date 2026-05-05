package auth

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterAuth struct {
}

func init() {
	log.Println("init auth router")
	ru := &RouterAuth{}
	router.Register(ru)
}

func (*RouterAuth) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/auth", h.list)
	// 以下接口需要 system.account.auth 节点权限
	adminGroup := group.Group("")
	adminGroup.Use(midd.NodeVerify("system.account.auth"))
	adminGroup.POST("/auth/add", h.add)
	adminGroup.POST("/auth/edit", h.edit)
	adminGroup.POST("/auth/apply", h.apply)
	adminGroup.POST("/auth/forbid", h.forbid)
	adminGroup.POST("/auth/resume", h.resume)
	adminGroup.POST("/auth/setDefault", h.setDefault)
	adminGroup.POST("/auth/del", h.del)
}

