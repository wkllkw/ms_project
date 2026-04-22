package invite_link

import (
	"log"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterInviteLink struct{}

func init() {
	log.Println("init invite_link router")
	ru := &RouterInviteLink{}
	router.Register(ru)
}

func (*RouterInviteLink) Route(r *gin.Engine) {
	h := New()
	group := r.Group("/project/invite_link")
	group.Use(midd.TokenVerify())
	group.POST("/save", h.createInvite)
	group.POST("/_read", h.getInviteDetail)
}
