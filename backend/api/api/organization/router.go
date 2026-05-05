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
	// 组织管理操作：需要 project.manage 节点权限
	manageGroup := group.Group("")
	manageGroup.Use(midd.NodeVerify("project.manage"))
	manageGroup.POST("/save", h.createOrganization)
	manageGroup.POST("/edit", h.editOrganization)
	manageGroup.POST("/delete", h.deleteOrganization)
}
