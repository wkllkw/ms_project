package task

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterTask struct {
}

func init() {
	log.Println("init task router")
	ru := &RouterTask{}
	router.Register(ru)
}

func (*RouterTask) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	group.POST("/task", h.list)
	group.POST("/task/selfList", h.selfList)
	group.POST("/task/getListByTaskTag", h.getListByTaskTag)
	group.POST("/task/taskSources", h.taskSources)
	group.POST("/task/sort", h.sort)
	group.POST("/task/save", h.save)
	group.POST("/task/edit", h.edit)
	group.POST("/task/taskToTags", h.taskToTags)
	group.POST("/task/setTag", h.setTag)
	group.POST("/task/like", h.like)
	group.POST("/task/star", h.star)
	group.POST("/task/createComment", h.createComment)
	group.POST("/task/assignTask", h.assignTask)
	group.POST("/task/batchAssignTask", h.batchAssignTask)
	group.POST("/task/read", h.read)
	group.POST("/task/taskDone", h.taskDone)
	group.POST("/task/setPrivate", h.setPrivate)
	group.POST("/task/recycle", h.recycle)
	group.POST("/task/recycleBatch", h.recycleBatch)
	group.POST("/task/recovery", h.recovery)
	group.POST("/task/delete", h.del)
	group.POST("/task/dateTotalForProject", h.dateTotalForProject)
	group.POST("/task/taskLog", h.taskLog)
	group.POST("/task/_taskWorkTimeList", h.taskWorkTimeList)
	group.POST("/task/saveTaskWorkTime", h.saveTaskWorkTime)
	group.POST("/task/editTaskWorkTime", h.editTaskWorkTime)
	group.POST("/task/delTaskWorkTime", h.delTaskWorkTime)
	group.POST("/task/_downloadTemplate", h.downloadTemplate)
	group.POST("/task/uploadFile", h.uploadFile)
}

