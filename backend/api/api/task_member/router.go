package task_member

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTaskMember struct{}

func init() {
	log.Println("init task_member router")
	ru := &RouterTaskMember{}
	router.Register(ru)
}

func (*RouterTaskMember) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task_member", h.list)
	group.POST("/task_member/searchInviteMember", h.searchInviteMember)
	group.POST("/task_member/inviteMember", h.inviteMember)
	group.POST("/task_member/inviteMemberBatch", h.inviteMemberBatch)
}
