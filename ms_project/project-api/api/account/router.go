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
	group.POST("/account/forbid", h.forbid)
	group.POST("/account/resume", h.resume)
	group.POST("/account/add", h.add)
	group.POST("/account/edit", h.edit)
	group.POST("/account/auth", h.auth)
	group.POST("/account/del", h.del)
	group.POST("/account/read", h.read)
	group.POST("/account/_syncDetail", h.syncDetail)
	group.POST("/account/_joinByInviteLink", h.joinByInviteLink)
}

