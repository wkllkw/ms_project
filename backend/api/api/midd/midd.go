package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/user/login"
	"time"
)

func TokenVerify() func(*gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		//1.从header中获取token
		token := c.GetHeader("Authorization")
		//2.调用user服务进行token认证
		ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		response, err := rpc.LoginServiceClient.TokenVerify(ctx, &login.LoginMessage{Token: token})
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}
		//3.处理结果，认证通过 将信息放入gin的上下文 失败返回未登录
		c.Set("memberId", response.Member.Id)
		c.Set("memberName", response.Member.Name)

		// 4. 处理组织切换：优先使用请求头的 organizationCode
		headerOrgCode := c.GetHeader("organizationCode")
		orgCode := response.Member.OrganizationCode // 默认使用 TokenVerify 返回的组织

		if headerOrgCode != "" && headerOrgCode != orgCode {
			orgResp, orgErr := rpc.LoginServiceClient.MyOrgList(ctx, &login.UserMessage{MemId: response.Member.Id})
			if orgErr == nil {
				for _, org := range orgResp.OrganizationList {
					if org.Code == headerOrgCode {
						orgCode = headerOrgCode
						break
					}
				}
			}
		}

		c.Set("organizationCode", orgCode)
		c.Next()
	}
}

// NodeVerify 节点权限校验中间件
// 使用方式：在路由上 .Use(midd.NodeVerify("system.menu")) 或 midd.NodeVerify("system.menu")
// 必须传入需要校验的节点名称，不再支持从请求头读取（防止伪造）
func NodeVerify(requiredNode ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		memberId := c.GetInt64("memberId")
		if memberId == 0 {
			c.JSON(http.StatusOK, result.Fail(401, "未登录"))
			c.Abort()
			return
		}

		// 确定要校验的节点：只接受中间件参数，不接受请求头
		node := ""
		if len(requiredNode) > 0 && requiredNode[0] != "" {
			node = requiredNode[0]
		}

		// 空节点或 "#" 表示不需要权限校验
		if node == "" || node == "#" {
			c.Next()
			return
		}

		db := gorms.GetDB()
		if !authz.HasNode(db, memberId, node) {
			c.JSON(http.StatusOK, result.Fail(4031, "无权限操作资源"))
			c.Abort()
			return
		}
		c.Next()
	}
}

