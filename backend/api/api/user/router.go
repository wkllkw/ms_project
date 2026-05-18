package user

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/api/rpc"
	"test.com/project-api/router"
)

type RouterUser struct {
}

func init() {
	log.Println("init user router")
	ru := &RouterUser{}
	router.Register(ru)
}

func (*RouterUser) Route(r *gin.Engine) {
	//初始化grpc的客户端连接
	rpc.InitRpcUserClient()
	h := New()
	r.POST("/project/login/getCaptcha", h.getCaptcha)
	r.POST("/project/login/register", h.register)
	r.POST("/project/login", h.login)
	r.POST("/project/login/_out", h._out)
	// 忘记密码（无需认证）
	r.POST("/project/login/_getMailCaptcha", h.getMailCaptcha)
	r.POST("/project/login/_resetPasswordByMail", h.resetPasswordByMail)
	// 需要认证的路由
	authGroup := r.Group("/project")
	authGroup.Use(midd.TokenVerify())
	authGroup.POST("/login/_bindMobile", h.bindMobile)
	authGroup.POST("/login/_bindMail", h.bindMail)
	authGroup.POST("/login/_unbindDingtalk", h.unbindDingtalk)
}
