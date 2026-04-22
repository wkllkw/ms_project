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
	group.POST("", h.getOrganizationList)
	group.POST("/save", h.createOrganization)
	group.POST("/edit", h.editOrganization)
	group.POST("/delete", h.deleteOrganization)
	group.POST("/_getOrgList", h.getOrgList)
	group.POST("/_quitOrganization", h.quitOrganization)
}
