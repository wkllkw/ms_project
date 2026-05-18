package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
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

// OrgNodeVerify 组织级节点权限校验中间件
// 使用方式：在路由上 .Use(midd.OrgNodeVerify("organization.setting"))
// 先通过 context 获取当前 organizationCode，再检查用户在该组织的权限
func OrgNodeVerify(requiredNode ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		memberId := c.GetInt64("memberId")
		if memberId == 0 {
			c.JSON(http.StatusOK, result.Fail(401, "未登录"))
			c.Abort()
			return
		}

		node := ""
		if len(requiredNode) > 0 && requiredNode[0] != "" {
			node = requiredNode[0]
		}

		if node == "" || node == "#" {
			c.Next()
			return
		}

		// 从 context 或请求参数中获取 organizationCode
		orgCodeStr, _ := c.Get("organizationCode")
		var orgCode int64
		if orgStr, ok := orgCodeStr.(string); ok && orgStr != "" {
			orgCode, _ = codecs.DecryptInt64(orgStr)
		}
		// fallback：从 PostForm 获取
		if orgCode == 0 {
			if orgPost := c.PostForm("organizationCode"); orgPost != "" {
				orgCode = decryptOrgCodeFromString(orgPost)
			}
		}

		if orgCode == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "缺少组织参数"))
			c.Abort()
			return
		}

		db := gorms.GetDB()
		if !authz.HasOrgNode(db, memberId, orgCode, node) {
			c.JSON(http.StatusOK, result.Fail(4031, "无权限操作组织资源"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func decryptOrgCodeFromString(s string) int64 {
	i, err := codecs.DecryptInt64(s)
	if err == nil && i > 0 {
		return i
	}
	// 尝试作为普通数字解析
	var num int64
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			num = num*10 + int64(ch-'0')
		} else {
			return 0
		}
	}
	return num
}

