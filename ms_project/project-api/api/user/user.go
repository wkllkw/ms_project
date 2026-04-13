package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model/user"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/user/login"
)

type HandlerUser struct {
}

func New() *HandlerUser {
	return &HandlerUser{}
}

func (*HandlerUser) getCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	mobile := ctx.PostForm("mobile")
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rsp, err := rpc.LoginServiceClient.GetCaptcha(c, &login.CaptchaMessage{Mobile: mobile})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	ctx.JSON(http.StatusOK, result.Success(rsp.Code))
}

func (u *HandlerUser) register(c *gin.Context) {
	//1.接收参数 参数模型
	result := &common.Result{}
	var req user.RegisterReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}
	//2.校验参数 判断参数是否合法
	if err := req.Verify(); err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	//3.调用user grpc服务 获取响应
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &login.RegisterMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}
	_, err = rpc.LoginServiceClient.Register(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	//4.返回结果
	c.JSON(http.StatusOK, result.Success(""))
}

func (u *HandlerUser) login(c *gin.Context) {
	//1.接收参数 参数模型
	result := &common.Result{}
	var req user.LoginReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}
	//2.调用user grpc 完成登录
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &login.LoginMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}
	loginRsp, err := rpc.LoginServiceClient.Login(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	rsp := &user.LoginRsp{}
	err = copier.Copy(rsp, loginRsp)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}
	//4.返回结果
	c.JSON(http.StatusOK, result.Success(rsp))
}

func (u *HandlerUser) myOrgList(c *gin.Context) {
	result := &common.Result{}
	memberIdStr, _ := c.Get("memberId")
	memberId := memberIdStr.(int64)
	list, err2 := rpc.LoginServiceClient.MyOrgList(context.Background(), &login.UserMessage{MemId: memberId})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	if list.OrganizationList == nil {
		c.JSON(http.StatusOK, result.Success([]*user.OrganizationList{}))
		return
	}
	var orgs []*user.OrganizationList
	copier.Copy(&orgs, list.OrganizationList)
	c.JSON(http.StatusOK, result.Success(orgs))
}

func (u *HandlerUser) editPersonal(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil || id != memberId {
		c.JSON(http.StatusOK, result.Fail(400, "code无效或权限不足"))
		return
	}
	updates := map[string]any{}
	for _, k := range []string{"name", "mobile", "email", "avatar", "description"} {
		if v := c.PostForm(k); v != "" {
			updates[k] = v
		}
	}
	if len(updates) == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "无更新字段"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Table("ms_member").Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (u *HandlerUser) editPassword(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil || id != memberId {
		c.JSON(http.StatusOK, result.Fail(400, "code无效或权限不足"))
		return
	}
	oldPassword := c.PostForm("password")
	newPassword := c.PostForm("newPassword")
	confirmPassword := c.PostForm("confirmPassword")
	if newPassword != confirmPassword {
		c.JSON(http.StatusOK, result.Fail(400, "两次输入的新密码不一致"))
		return
	}
	if len(newPassword) < 6 {
		c.JSON(http.StatusOK, result.Fail(400, "新密码长度至少6位"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 验证旧密码（MD5加密后比对）
	var storedPwd string
	if err := db.Table("ms_member").Where("id=?", id).Select("password").Scan(&storedPwd).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	oldMd5 := fmt.Sprintf("%x", md5.Sum([]byte(oldPassword)))
	if storedPwd != oldMd5 {
		c.JSON(http.StatusOK, result.Fail(400, "旧密码不正确"))
		return
	}
	newMd5 := fmt.Sprintf("%x", md5.Sum([]byte(newPassword)))
	if err := db.Table("ms_member").Where("id=?", id).Update("password", newMd5).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "密码更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (u *HandlerUser) bindMobile(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	mobile := c.PostForm("mobile")
	captcha := c.PostForm("captcha")
	if mobile == "" || captcha == "" {
		c.JSON(http.StatusOK, result.Fail(400, "手机号和验证码不能为空"))
		return
	}
	// 从 Redis 验证验证码
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rsp, err := rpc.LoginServiceClient.GetCaptcha(ctx, &login.CaptchaMessage{Mobile: mobile})
	if err == nil && rsp != nil {
		if rsp.Code != captcha {
			c.JSON(http.StatusOK, result.Fail(400, "验证码不正确"))
			return
		}
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Table("ms_member").Where("id=?", memberId).Update("mobile", mobile).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "绑定手机失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (u *HandlerUser) bindMail(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	mail := c.PostForm("mail")
	if mail == "" {
		c.JSON(http.StatusOK, result.Fail(400, "邮箱不能为空"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Table("ms_member").Where("id=?", memberId).Update("email", mail).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "绑定邮箱失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (u *HandlerUser) unbindDingtalk(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 假设钉钉unionid存储在ms_member表的dingtalk_unionid字段
	if err := db.Table("ms_member").Where("id=?", memberId).Update("dingtalk_unionid", "").Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "解绑钉钉失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}
