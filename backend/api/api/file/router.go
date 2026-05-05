package file

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/router"
)

type RouterFile struct {
}

func init() {
	log.Println("init file router")
	ru := &RouterFile{}
	router.Register(ru)
}

func (*RouterFile) Route(r *gin.Engine) {
	group := r.Group("/project")
	group.Use(midd.TokenVerify())
	h := New()
	// 只读操作：需要 project.list 权限（通过菜单过滤控制）
	group.POST("/file", h.list)
	group.POST("/file/read", h.read)
	// 文件上传：需要 file:upload 节点
	group.POST("/file/uploadFiles", midd.NodeVerify("file:upload"), h.uploadFiles)
	// 文件编辑：需要 file:upload 节点（编辑文件信息属于上传权限范围）
	group.POST("/file/edit", midd.NodeVerify("file:upload"), h.edit)
	// 文件删除/回收：需要 file:delete 节点
	group.POST("/file/recycle", midd.NodeVerify("file:delete"), h.recycle)
	group.POST("/file/recovery", midd.NodeVerify("file:delete"), h.recovery)
	group.POST("/file/delete", midd.NodeVerify("file:delete"), h.del)
}
