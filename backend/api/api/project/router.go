package project

import (
	"log"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterProject struct {
}

func init() {
	log.Println("init project router")
	ru := &RouterProject{}
	router.Register(ru)
}

func (*RouterProject) Route(r *gin.Engine) {
	//初始化grpc的客户端连接
	InitRpcProjectClient()
	h := New()
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	// 只读操作
	group.POST("/project/selfList", h.myProjectList)
	group.POST("/project", h.myProjectList)
	group.POST("/project/read", h.readProject)
	group.POST("/project/analysis", h.analysis)
	group.POST("/project/_projectStats", h.projectStats)
	group.POST("/project/_getProjectReport", h.projectReport)
	group.POST("/project/getLogBySelfProject", h.getLogBySelfProject)
	group.POST("/project/accountList", h.accountList)
	group.POST("/project/taskSelfList", h.taskSelfList)
	group.POST("/project/eventsMyList", h.eventsMyList)
	group.POST("/project/notifyNoReads", h.notifyNoReads)
	// 项目创建：需要 project.list 权限（已登录用户可创建项目）
	group.POST("/project/save", h.projectSave)
	group.POST("/project/uploadCover", h.uploadCover)
	// 项目管理操作：需要 project.manage 节点权限
	manageGroup := group.Group("")
	manageGroup.Use(midd.NodeVerify("project.manage"))
	manageGroup.POST("/project/edit", h.editProject)
	manageGroup.POST("/project/recycle", h.recycleProject)
	manageGroup.POST("/project/recovery", h.recoveryProject)
	manageGroup.POST("/project/archive", h.archiveProject)
	manageGroup.POST("/project/recoveryArchive", h.recoveryArchiveProject)
	manageGroup.POST("/project/saveAsTemplate", h.saveAsTemplate)
	// 项目删除：需要 project.delete 节点权限
	deleteGroup := group.Group("")
	deleteGroup.Use(midd.NodeVerify("project.delete"))
	deleteGroup.POST("/project/delete", h.deleteProject)
	deleteGroup.POST("/project/batchDelete", h.batchDeleteProjects)
	deleteGroup.POST("/project/batchRecovery", h.batchRecoveryProjects)
}
