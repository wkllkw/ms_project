package file

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterFile struct {
}

func init() {
	log.Println("init file router")
	ru := &RouterFile{}
	router.Register(ru)
}

func (*RouterFile) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/file", h.list)
	group.POST("/file/read", h.read)
	group.POST("/file/edit", h.edit)
	group.POST("/file/uploadFiles", h.uploadFiles)
	group.POST("/file/recycle", h.recycle)
	group.POST("/file/recovery", h.recovery)
	group.POST("/file/delete", h.del)
}
