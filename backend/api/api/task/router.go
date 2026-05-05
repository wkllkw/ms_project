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
	// 只读操作：需要 project.list 权限（通过菜单过滤控制）
	group.POST("/task", h.list)
	group.POST("/task/selfList", h.selfList)
	group.POST("/task/getListByTaskTag", h.getListByTaskTag)
	group.POST("/task/taskSources", h.taskSources)
	group.POST("/task/read", h.read)
	group.POST("/task/dateTotalForProject", h.dateTotalForProject)
	group.POST("/task/taskLog", h.taskLog)
	group.POST("/task/_taskWorkTimeList", h.taskWorkTimeList)
	group.POST("/task/_downloadTemplate", h.downloadTemplate)
	// 任务创建：需要 task:create 节点
	group.POST("/task/save", midd.NodeVerify("task:create"), h.save)
	// 任务编辑：需要 task:edit 节点
	group.POST("/task/edit", midd.NodeVerify("task:edit"), h.edit)
	group.POST("/task/sort", midd.NodeVerify("task:edit"), h.sort)
	group.POST("/task/taskToTags", midd.NodeVerify("task:edit"), h.taskToTags)
	group.POST("/task/setTag", midd.NodeVerify("task:edit"), h.setTag)
	group.POST("/task/taskDone", midd.NodeVerify("task:edit"), h.taskDone)
	group.POST("/task/batchDone", midd.NodeVerify("task:edit"), h.batchDone)
	group.POST("/task/setPrivate", midd.NodeVerify("task:edit"), h.setPrivate)
	group.POST("/task/recovery", midd.NodeVerify("task:edit"), h.recovery)
	group.POST("/task/like", midd.NodeVerify("task:edit"), h.like)
	group.POST("/task/star", midd.NodeVerify("task:edit"), h.star)
	group.POST("/task/createComment", midd.NodeVerify("task:edit"), h.createComment)
	group.POST("/task/saveTaskWorkTime", midd.NodeVerify("task:edit"), h.saveTaskWorkTime)
	group.POST("/task/editTaskWorkTime", midd.NodeVerify("task:edit"), h.editTaskWorkTime)
	group.POST("/task/uploadFile", midd.NodeVerify("task:edit"), h.uploadFile)
	// 任务分配：需要 task:assign 节点
	group.POST("/task/assignTask", midd.NodeVerify("task:assign"), h.assignTask)
	group.POST("/task/batchAssignTask", midd.NodeVerify("task:assign"), h.batchAssignTask)
	// 任务删除：需要 task:delete 节点
	group.POST("/task/recycle", midd.NodeVerify("task:delete"), h.recycle)
	group.POST("/task/recycleBatch", midd.NodeVerify("task:delete"), h.recycleBatch)
	group.POST("/task/delete", midd.NodeVerify("task:delete"), h.del)
	group.POST("/task/delTaskWorkTime", midd.NodeVerify("task:delete"), h.delTaskWorkTime)
}

