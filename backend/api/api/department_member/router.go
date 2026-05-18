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
	h := New()

	// 只读操作：只需要 Token 验证
	readGroup := r.Group("/project")
	readGroup.Use(midd.TokenVerify())
	readGroup.POST("/department_member/searchInviteMember", h.searchInviteMember)
	readGroup.POST("/department_member/index", h.index)
	readGroup.POST("/department_member/detail", h.detail)

	// 写操作：需要组织部门管理权限
	writeGroup := r.Group("/project")
	writeGroup.Use(midd.TokenVerify(), midd.OrgNodeVerify("organization.department"))
	writeGroup.POST("/department_member/inviteMember", h.inviteMember)
	writeGroup.POST("/department_member/removeMember", h.removeMember)
}

