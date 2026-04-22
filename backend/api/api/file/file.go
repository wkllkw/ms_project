package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerFile struct {
}

func New() *HandlerFile {
	return &HandlerFile{}
}

type fileRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	MemberCode  int64  `gorm:"column:member_code"`
	Title       string `gorm:"column:title"`
	FileName    string `gorm:"column:file_name"`
	FileType    string `gorm:"column:file_type"`
	FileSize    int64  `gorm:"column:file_size"`
	FileUrl     string `gorm:"column:file_url"`
	FilePath    string `gorm:"column:file_path"`
	Description string `gorm:"column:description"`
	Deleted     int8   `gorm:"column:deleted"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*fileRow) TableName() string { return "ms_file" }

// list 获取文件列表
func (h *HandlerFile) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var total int64
	query := db.Model(&fileRow{}).Where("project_code=? AND deleted=0", pid)

	// 搜索关键字
	keyword := c.PostForm("keyword")
	if keyword != "" {
		query = query.Where("title LIKE ? OR file_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	_ = query.Count(&total).Error

	var rows []fileRow
	_ = query.Order("id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Find(&rows).Error

	// 获取成员信息
	memberIds := make([]int64, 0)
	for _, r := range rows {
		memberIds = append(memberIds, r.MemberCode)
	}
	memberMap := make(map[int64]memberInfo)
	if len(memberIds) > 0 {
		var members []memberInfo
		db.Table("ms_member").Where("id IN ?", memberIds).Find(&members)
		for _, m := range members {
			memberMap[m.Id] = m
		}
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		member := memberMap[r.MemberCode]
		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"title":       r.Title,
			"fullName":    r.FileName,
			"extension":   getFileExtension(r.FileName),
			"fileType":    r.FileType,
			"size":        r.FileSize,
			"file_url":    r.FileUrl,
			"filePath":    r.FilePath,
			"description": r.Description,
			"create_time": r.CreateTime,
			"creatorName": member.Name,
			"memberCode":  codecs.EncryptInt64(r.MemberCode),
			"projectCode": projectCode,
		})
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

type memberInfo struct {
	Id     int64  `gorm:"column:id"`
	Name   string `gorm:"column:name"`
	Avatar string `gorm:"column:avatar"`
}

// read 获取文件详情
func (h *HandlerFile) read(c *gin.Context) {
	result := &common.Result{}
	fileCode := c.PostForm("fileCode")

	if fileCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(fileCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var file fileRow
	if err := db.Where("id=? AND deleted=0", fid).First(&file).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "文件不存在"))
		return
	}

	// 获取成员信息
	var member memberInfo
	db.Table("ms_member").Where("id=?", file.MemberCode).First(&member)

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        codecs.EncryptInt64(file.Id),
		"title":       file.Title,
		"fullName":    file.FileName,
		"fileType":    file.FileType,
		"fileSize":    file.FileSize,
		"fileUrl":     file.FileUrl,
		"filePath":    file.FilePath,
		"description": file.Description,
		"createTime":  file.CreateTime,
		"createBy":    member.Name,
		"memberCode":  codecs.EncryptInt64(file.MemberCode),
		"projectCode": codecs.EncryptInt64(file.ProjectCode),
	}))
}

// edit 编辑文件信息
func (h *HandlerFile) edit(c *gin.Context) {
	result := &common.Result{}
	fileCode := c.PostForm("fileCode")
	title := c.PostForm("title")
	description := c.PostForm("description")

	if fileCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(fileCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"update_time": time.Now().UnixMilli(),
	}
	if title != "" {
		updates["title"] = title
	}
	if description != "" {
		updates["description"] = description
	}

	if err := db.Model(&fileRow{}).Where("id=?", fid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": fileCode}))
}

// uploadFiles 上传文件
// @Summary 上传文件
// @Description 上传文件到项目，支持图片/文档/视频等类型
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Param projectCode formData string true "项目加密ID"
// @Success 200 {object} common.Result "返回文件信息(code/title/fileType/fileSize/fileUrl)"
// @Failure 400 {object} common.Result "参数错误或文件过大"
// @Security ApiKeyAuth
// @Router /file/upload [post]
func (h *HandlerFile) uploadFiles(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "请选择要上传的文件"))
		return
	}
	defer file.Close()

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileType := getFileType(ext)

	// 创建上传目录
	uploadDir := filepath.Join("uploads", "files", time.Now().Format("2006/01/02"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建上传目录失败"))
		return
	}

	// 生成文件名
	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), memberId, ext)
	filePath := filepath.Join(uploadDir, fileName)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建文件失败"))
		return
	}
	defer dst.Close()

	// 复制文件内容
	fileSize, err := io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "保存文件失败"))
		return
	}

	// 生成文件URL
	fileUrl := "/" + strings.ReplaceAll(filePath, "\\", "/")

	// 获取文件标题（不带扩展名的文件名）
	title := strings.TrimSuffix(header.Filename, ext)

	// 保存到数据库
	db := gorms.GetDB().WithContext(c.Request.Context())
	now := time.Now().UnixMilli()
	row := &fileRow{
		ProjectCode: pid,
		MemberCode:  memberId,
		Title:       title,
		FileName:    header.Filename,
		FileType:    fileType,
		FileSize:    fileSize,
		FileUrl:     fileUrl,
		FilePath:    filePath,
		CreateTime:  now,
		UpdateTime:  now,
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "保存文件信息失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":     codecs.EncryptInt64(row.Id),
		"title":    row.Title,
		"fullName": row.FileName,
		"fileType": row.FileType,
		"fileSize": row.FileSize,
		"fileUrl":  row.FileUrl,
	}))
}

// recycle 移动到回收站
func (h *HandlerFile) recycle(c *gin.Context) {
	result := &common.Result{}
	fileCode := c.PostForm("fileCode")

	if fileCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(fileCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Model(&fileRow{}).Where("id=?", fid).Updates(map[string]any{
		"deleted":     1,
		"update_time": time.Now().UnixMilli(),
	}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "操作失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": fileCode}))
}

// recovery 从回收站恢复
func (h *HandlerFile) recovery(c *gin.Context) {
	result := &common.Result{}
	fileCode := c.PostForm("fileCode")

	if fileCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(fileCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Model(&fileRow{}).Where("id=?", fid).Updates(map[string]any{
		"deleted":     0,
		"update_time": time.Now().UnixMilli(),
	}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "操作失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": fileCode}))
}

// del 彻底删除文件
func (h *HandlerFile) del(c *gin.Context) {
	result := &common.Result{}
	fileCode := c.PostForm("fileCode")

	if fileCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(fileCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "fileCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取文件信息
	var file fileRow
	if err := db.Where("id=?", fid).First(&file).Error; err == nil {
		// 删除物理文件
		if file.FilePath != "" {
			os.Remove(file.FilePath)
		}
	}

	// 删除数据库记录
	db.Delete(&fileRow{}, fid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": fileCode}))
}

// getFileType 获取文件类型
func getFileType(ext string) string {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	docExts := []string{".doc", ".docx", ".pdf", ".txt", ".xls", ".xlsx", ".ppt", ".pptx"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg"}
	codeExts := []string{".js", ".ts", ".vue", ".go", ".java", ".py", ".c", ".cpp", ".h", ".css", ".html", ".json", ".xml", ".sql", ".sh"}

	for _, e := range imageExts {
		if ext == e {
			return "image"
		}
	}
	for _, e := range docExts {
		if ext == e {
			return "doc"
		}
	}
	for _, e := range videoExts {
		if ext == e {
			return "video"
		}
	}
	for _, e := range audioExts {
		if ext == e {
			return "audio"
		}
	}
	for _, e := range codeExts {
		if ext == e {
			return "code"
		}
	}
	return "other"
}

// getFileExtension 获取文件扩展名（不含点）
func getFileExtension(filename string) string {
	if filename == "" {
		return ""
	}
	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		return ext[1:] // 去掉点
	}
	return ""
}

// strconv import
var _ = strconv.Itoa
