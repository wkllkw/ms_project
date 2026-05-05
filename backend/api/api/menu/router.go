package menu

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterMenu struct {
}

func init() {
	log.Println("init menu router")
	ru := &RouterMenu{}
	router.Register(ru)
}

func (*RouterMenu) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/menu/menu", h.menu)
	// 菜单增删改需要 system.menu 节点权限
	adminGroup := group.Group("")
	adminGroup.Use(midd.NodeVerify("system.menu"))
	adminGroup.POST("/menu/menuAdd", h.menuAdd)
	adminGroup.POST("/menu/menuEdit", h.menuEdit)
	adminGroup.POST("/menu/menuDel", h.menuDel)
	adminGroup.POST("/menu/menuForbid", h.menuForbid)
	adminGroup.POST("/menu/menuResume", h.menuResume)
}

