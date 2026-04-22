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
	group.POST("/menu/menuAdd", h.menuAdd)
	group.POST("/menu/menuEdit", h.menuEdit)
	group.POST("/menu/menuDel", h.menuDel)
	group.POST("/menu/menuForbid", h.menuForbid)
	group.POST("/menu/menuResume", h.menuResume)
}

