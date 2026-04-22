package department_member

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterDepartmentMember struct {
}

func init() {
	log.Println("init department_member router")
	ru := &RouterDepartmentMember{}
	router.Register(ru)
}

func (*RouterDepartmentMember) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/department_member/searchInviteMember", h.searchInviteMember)
	group.POST("/department_member/inviteMember", h.inviteMember)
	group.POST("/department_member/removeMember", h.removeMember)
	group.POST("/department_member/index", h.index)
	group.POST("/department_member/detail", h.detail)
}

