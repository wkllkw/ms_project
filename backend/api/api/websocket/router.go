package websocket

import (
	"github.com/gin-gonic/gin"
	"test.com/project-api/router"
)

type RouterWebSocket struct{}

func init() {
	router.Register(&RouterWebSocket{})
}

func (r *RouterWebSocket) Route(engine *gin.Engine) {
	engine.GET("/ws", HandleWebSocket)
}
