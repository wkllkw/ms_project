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
	group.POST("/index", h.index)
	group.POST("/project/selfList", h.myProjectList)
	group.POST("/project", h.myProjectList)
	group.POST("/project_template", h.projectTemplate)
	group.POST("/project_template/delete", h.projectTemplateDelete)
	group.POST("/project_template/save", h.projectTemplateSave)
	group.POST("/project_template/edit", h.projectTemplateSave)
	group.POST("/project_template/uploadCover", h.uploadTemplateCover)
	group.POST("/project/save", h.projectSave)
	group.POST("/project/read", h.readProject)
	group.POST("/project/recycle", h.recycleProject)
	group.POST("/project/recovery", h.recoveryProject)
	group.POST("/project/quit", h.quitProject)
	group.POST("/project/archive", h.archiveProject)
	group.POST("/project/recoveryArchive", h.recoveryArchiveProject)
	group.POST("/project/delete", h.deleteProject)
	group.POST("/project/batchDelete", h.batchDeleteProjects)
	group.POST("/project/batchRecovery", h.batchRecoveryProjects)
	group.POST("/project/analysis", h.analysis)
	group.POST("/project/_projectStats", h.projectStats)
	group.POST("/project/_getProjectReport", h.projectReport)
	group.POST("/project_collect/collect", h.collectProject)
	group.POST("/project/edit", h.editProject)
	group.POST("/project/saveAsTemplate", h.saveAsTemplate)
	group.POST("/project/getLogBySelfProject", h.getLogBySelfProject)
	group.POST("/project/accountList", h.accountList)
	group.POST("/project/taskSelfList", h.taskSelfList)
	group.POST("/project/eventsMyList", h.eventsMyList)
	group.POST("/project/notifyNoReads", h.notifyNoReads)
}
