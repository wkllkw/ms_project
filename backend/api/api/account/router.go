package account

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterAccount struct {
}

func init() {
	log.Println("init account router")
	ru := &RouterAccount{}
	router.Register(ru)
}

func (*RouterAccount) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/account", h.list)
	group.POST("/account/_allList", h.allList)
	group.POST("/account/read", h.read)
	group.POST("/account/_syncDetail", h.syncDetail)
	group.POST("/account/_joinByInviteLink", h.joinByInviteLink)
	// 账号增删改禁启用需要 system.account 节点权限
	adminGroup := group.Group("")
	adminGroup.Use(midd.NodeVerify("system.account"))
	adminGroup.POST("/account/forbid", h.forbid)
	adminGroup.POST("/account/resume", h.resume)
	adminGroup.POST("/account/add", h.add)
	adminGroup.POST("/account/edit", h.edit)
	adminGroup.POST("/account/auth", h.auth)
	adminGroup.POST("/account/del", h.del)
}

