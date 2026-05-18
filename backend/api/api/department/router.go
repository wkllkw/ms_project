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
	h := New()

	// 只读操作：只需要 Token 验证
	readGroup := r.Group("/project")
	readGroup.Use(midd.TokenVerify())
	readGroup.POST("/department", h.list)
	readGroup.POST("/department/read", h.read)

	// 写操作：需要组织部门管理权限
	writeGroup := r.Group("/project")
	writeGroup.Use(midd.TokenVerify(), midd.OrgNodeVerify("organization.department"))
	writeGroup.POST("/department/save", h.save)
	writeGroup.POST("/department/edit", h.edit)
	writeGroup.POST("/department/delete", h.del)
}

