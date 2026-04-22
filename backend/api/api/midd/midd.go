package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"test.com/project-api/api/rpc"
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

