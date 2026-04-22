package project_member

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectMember struct {
}

func init() {
	log.Println("init project_member router")
	ru := &RouterProjectMember{}
	router.Register(ru)
}

func (*RouterProjectMember) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/project_member/index", h.index)
	group.POST("/project_member/searchInviteMember", h.searchInviteMember)
	group.POST("/project_member/inviteMember", h.inviteMember)
	group.POST("/project_member/removeMember", h.removeMember)
	group.POST("/project_member/_listForInvite", h.listForInvite)
	group.POST("/project_member/_joinByInviteLink", h.joinByInviteLink)
}

