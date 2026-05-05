package index

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterIndex struct{}

func init() {
	log.Println("init index router")
	ru := &RouterIndex{}
	router.Register(ru)
}

func (*RouterIndex) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/index", h.index)
	group.POST("/index/_menus", h.menus)
	group.POST("/index/nodes", h.nodes)
	group.POST("/index/changeCurrentOrganization", h.changeCurrentOrganization)
	group.POST("/index/systemConfig", h.systemConfig)
	group.POST("/index/info", h.info)
	group.POST("/index/editPersonal", h.editPersonal)
	group.POST("/index/editPassword", h.editPassword)
	group.POST("/index/uploadAvatar", h.uploadAvatar)
	group.POST("/index/uploadImg", h.uploadImg)
	group.POST("/index/bindClientId", h.bindClientId)
}
