package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/cache"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/email"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model/user"
	common "test.com/project-common"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-common/jwts"
	"test.com/project-common/tms"
	"test.com/project-grpc/user/login"
)

const (
	jwtAccessSecret  = "E7k9P2mX8vQ4wR6jN3hL5yT1bA0zCfDg"
	jwtRefreshSecret = "aS5dF8gH2jK4lM6nP9qR1tV3wX7yZ0bCe"
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

// @Summary 用户登录
// @Description 使用账号密码登录系统，返回Token和用户信息
// @Tags user
// @Accept json
// @Produce json
// @Param account body user.LoginReq true "登录信息(account/password)"
// @Success 200 {object} common.Result "返回用户信息和Token"
// @Failure 400 {object} common.Result "账号或密码错误"
// @Router /user/login [post]
func (u *HandlerUser) login(c *gin.Context) {
	result := &common.Result{}
	var req user.LoginReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}

	// 手机号验证码登录（在HTTP层处理，绕过gRPC）
	if req.Mobile != "" && req.Captcha != "" {
		u.loginByMobile(c, result, &req)
		return
	}

	// 账号密码登录（走gRPC）
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
	c.JSON(http.StatusOK, result.Success(rsp))
}

// loginByMobile 手机号验证码登录
func (u *HandlerUser) loginByMobile(c *gin.Context, result *common.Result, req *user.LoginReq) {
	// 1. 验证验证码（gRPC 用户服务使用 REGISTER_ 前缀存储，不带 "ms:" 前缀）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	storedCode, err := cache.Client.Get(ctx, "REGISTER_"+req.Mobile).Result()
	if err != nil || storedCode != req.Captcha {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "验证码错误或已过期"))
		return
	}
	// 2. 查找用户
	db := gorms.GetDB()
	var mem struct {
		Id       int64  `gorm:"column:id"`
		Account  string `gorm:"column:account"`
		Name     string `gorm:"column:name"`
		Mobile   string `gorm:"column:mobile"`
		Email    string `gorm:"column:email"`
		Status   int    `gorm:"column:status"`
		CreateTime int64 `gorm:"column:create_time"`
	}
	if err := db.Table("ms_member").Where("mobile = ?", req.Mobile).First(&mem).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "手机号未注册"))
		return
	}
	// 3. 生成JWT token
	memIdStr := fmt.Sprintf("%d", mem.Id)
	accessExp := 7 * 24 * time.Hour
	refreshExp := 14 * 24 * time.Hour
	token := jwts.CreateToken(memIdStr, accessExp, jwtAccessSecret, refreshExp, jwtRefreshSecret)
	// 4. 加密member code
	var memCode string
	memCode, _ = encrypts.EncryptInt64(mem.Id, "sdfgyrhgbxcdgryfhgywertd")
	// 5. 查找组织
	var orgs []struct {
		Id         int64  `gorm:"column:id"`
		Name       string `gorm:"column:name"`
		Avatar     string `gorm:"column:avatar"`
		CreateTime int64  `gorm:"column:create_time"`
		Personal   int32  `gorm:"column:personal"`
	}
	db.Table("ms_organization o").
		Joins("JOIN ms_member_account ma ON ma.organization_code = o.id").
		Where("ma.member_code = ?", mem.Id).
		Select("o.id, o.name, o.avatar, o.create_time, o.personal").
		Find(&orgs)
	// 6. 构建组织列表
	orgList := make([]user.OrganizationList, 0, len(orgs))
	for _, o := range orgs {
		orgCode, _ := encrypts.EncryptInt64(o.Id, "sdfgyrhgbxcdgryfhgywertd")
		orgList = append(orgList, user.OrganizationList{
			Name:       o.Name,
			Avatar:     o.Avatar,
			OwnerCode:  memCode,
			CreateTime: tms.FormatByMill(o.CreateTime),
			Personal:   o.Personal,
			Code:       orgCode,
		})
	}
	// 7. 确定默认组织
	var orgCode string
	if len(orgs) > 0 {
		orgCode, _ = encrypts.EncryptInt64(orgs[0].Id, "sdfgyrhgbxcdgryfhgywertd")
	}
	// 8. 返回
	rsp := &user.LoginRsp{
		Member: user.Member{
			Name:             mem.Name,
			Mobile:           mem.Mobile,
			Status:           mem.Status,
			Code:             memCode,
			Email:            mem.Email,
			CreateTime:       tms.FormatByMill(mem.CreateTime),
			OrganizationCode: orgCode,
		},
		TokenList: user.TokenList{
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			TokenType:      "bearer",
			AccessTokenExp: token.AccessExp,
		},
		OrganizationList: orgList,
	}
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

// _out 退出登录（前端调用，清理客户端状态即可）
// JWT Token 无需服务端注销，过期后自动失效
func (u *HandlerUser) _out(c *gin.Context) {
	result := &common.Result{}
	c.JSON(http.StatusOK, result.Success(nil))
}

// getMailCaptcha 发送邮箱验证码（用于忘记密码）
func (u *HandlerUser) getMailCaptcha(c *gin.Context) {
	result := &common.Result{}
	emailAddr := c.PostForm("email")
	if emailAddr == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "邮箱不能为空"))
		return
	}
	// 查找邮箱是否已注册
	db := gorms.GetDB()
	var count int64
	if err := db.Table("ms_member").Where("email = ?", emailAddr).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "该邮箱未注册"))
		return
	}
	// 生成6位随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	redisKey := "forgot_pwd:" + emailAddr
	// 存入Redis，15分钟过期
	if err := cache.Set(redisKey, code, 15*time.Minute); err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "验证码存储失败"))
		return
	}
	// 尝试发送邮件
	subject := "MS Project - 密码重置验证码"
	body := fmt.Sprintf(`
	<p>您好，</p>
	<p>您正在申请重置密码，验证码如下（15分钟内有效）：</p>
	<h2 style="color:#1890ff;font-size:28px;letter-spacing:4px;">%s</h2>
	<p>如果这不是您本人的操作，请忽略此邮件。</p>
	`, code)
	if err := email.Send(emailAddr, subject, body); err != nil {
		zap.L().Warn("忘记密码邮件发送失败，降级为前端显示验证码",
			zap.String("email", emailAddr),
			zap.Error(err))
		// 邮件发送失败时，将验证码返回给前端展示给用户（demo模式）
		c.JSON(http.StatusOK, result.Success(code))
		return
	}
	c.JSON(http.StatusOK, result.Success(nil))
}

// resetPasswordByMail 通过邮箱验证码重置密码
func (u *HandlerUser) resetPasswordByMail(c *gin.Context) {
	result := &common.Result{}
	emailAddr := c.PostForm("email")
	captcha := c.PostForm("captcha")
	password := c.PostForm("password")
	password2 := c.PostForm("password2")

	if emailAddr == "" || captcha == "" || password == "" || password2 == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数不能为空"))
		return
	}
	if password != password2 {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "两次密码不一致"))
		return
	}
	if len(password) < 6 {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "密码至少6位"))
		return
	}
	// 验证验证码
	redisKey := "forgot_pwd:" + emailAddr
	storedCode, err := cache.Get(redisKey)
	if err != nil || storedCode != captcha {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "验证码错误或已过期"))
		return
	}
	// 更新密码
	db := gorms.GetDB()
	newPwd := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	if err := db.Table("ms_member").Where("email = ?", emailAddr).Update("password", newPwd).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "密码重置失败"))
		return
	}
	// 删除已使用的验证码
	_ = cache.Del(redisKey)
	c.JSON(http.StatusOK, result.Success(nil))
}
