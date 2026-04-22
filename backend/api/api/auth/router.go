package auth

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterAuth struct {
}

func init() {
	log.Println("init auth router")
	ru := &RouterAuth{}
	router.Register(ru)
}

func (*RouterAuth) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/auth", h.list)
	group.POST("/auth/add", h.add)
	group.POST("/auth/edit", h.edit)
	group.POST("/auth/apply", h.apply)
	group.POST("/auth/forbid", h.forbid)
	group.POST("/auth/resume", h.resume)
	group.POST("/auth/setDefault", h.setDefault)
	group.POST("/auth/del", h.del)
}

