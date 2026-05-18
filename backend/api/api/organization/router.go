package organization

import (
	"log"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterOrganization struct{}

func init() {
	log.Println("init organization router")
	ru := &RouterOrganization{}
	router.Register(ru)
}

func (*RouterOrganization) Route(r *gin.Engine) {
	h := New()
	group := r.Group("/project/organization")
	group.Use(midd.TokenVerify())
	// 只读操作
	group.POST("", h.getOrganizationList)
	group.POST("/_getOrgList", h.getOrgList)
	group.POST("/_quitOrganization", h.quitOrganization)

	// 组织设置操作：需要 organization.setting 节点权限
	settingGroup := group.Group("")
	settingGroup.Use(midd.OrgNodeVerify("organization.setting"))
	settingGroup.POST("/save", h.createOrganization)
	settingGroup.POST("/edit", h.editOrganization)
	settingGroup.POST("/delete", h.deleteOrganization)

	// 组织成员管理：需要 organization.member 节点权限
	memberGroup := group.Group("")
	memberGroup.Use(midd.OrgNodeVerify("organization.member"))
	memberGroup.POST("/_listMembers", h.listMembersWithAuth)
	memberGroup.POST("/_setMemberAuth", h.setMemberAuth)
	memberGroup.POST("/_removeMemberAuth", h.removeMemberAuth)
	memberGroup.POST("/_getMemberAuth", h.getMemberAuth)
}
