package project_version

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProjectVersion struct {
}

func init() {
	log.Println("init project_version router")
	ru := &RouterProjectVersion{}
	router.Register(ru)
}

func (*RouterProjectVersion) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/project_version", h.list)
	group.POST("/project_version/save", h.save)
	group.POST("/project_version/edit", h.edit)
	group.POST("/project_version/delete", h.del)
	group.POST("/project_version/changeStatus", h.changeStatus)
	group.POST("/project_version/read", h.read)
	group.POST("/project_version/_getVersionTask", h.getVersionTask)
	group.POST("/project_version/_getVersionLog", h.getVersionLog)
	group.POST("/project_version/addVersionTask", h.addVersionTask)
	group.POST("/project_version/removeVersionTask", h.removeVersionTask)
}
