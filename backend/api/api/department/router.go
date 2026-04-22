package department

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterDepartment struct {
}

func init() {
	log.Println("init department router")
	ru := &RouterDepartment{}
	router.Register(ru)
}

func (*RouterDepartment) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/department", h.list)
	group.POST("/department/read", h.read)
	group.POST("/department/save", h.save)
	group.POST("/department/edit", h.edit)
	group.POST("/department/delete", h.del)
}

