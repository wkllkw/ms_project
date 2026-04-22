package events

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterEvents struct {
}

func init() {
	log.Println("init events router")
	ru := &RouterEvents{}
	router.Register(ru)
}

func (*RouterEvents) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/events", h.list)
	group.POST("/events/myList", h.myList)
	group.POST("/events/_getEventsLog", h.getEventsLog)
	group.POST("/events/read", h.read)
	group.POST("/events/confirmJoin", h.confirmJoin)
	group.POST("/events/getEventsListByCalendar", h.getEventsListByCalendar)
	group.POST("/events/save", h.save)
	group.POST("/events/edit", h.edit)
	group.POST("/events/delete", h.del)
}

