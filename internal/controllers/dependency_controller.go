package controllers

import (
	"strconv"
	"strings"

	"github.com/engigu/baihu-panel/internal/models"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

type DependencyController struct {
	service *services.DependencyService
}

func NewDependencyController() *DependencyController {
	return &DependencyController{
		service: services.NewDependencyService(),
	}
}

// List 获取依赖列表
func (c *DependencyController) List(ctx *gin.Context) {
	depType := ctx.Query("type")
	deps, err := c.service.List(depType)
	if err != nil {
		utils.ServerError(ctx, "获取依赖列表失败")
		return
	}
	utils.Success(ctx, deps)
}

// Create 添加依赖
func (c *DependencyController) Create(ctx *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Version string `json:"version"`
		Type    string `json:"type" binding:"required"`
		Remark  string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "参数错误")
		return
	}

	if req.Type != "py" && req.Type != "node" {
		utils.BadRequest(ctx, "类型必须是 py 或 node")
		return
	}

	dep := &models.Dependency{
		Name:    req.Name,
		Version: req.Version,
		Type:    req.Type,
		Remark:  req.Remark,
	}

	if err := c.service.Create(dep); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	utils.Success(ctx, dep)
}

// Delete 删除依赖
func (c *DependencyController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.BadRequest(ctx, "无效的 ID")
		return
	}

	if err := c.service.Delete(id); err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessMsg(ctx, "删除成功")
}

// Install 安装依赖
func (c *DependencyController) Install(ctx *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Version string `json:"version"`
		Type    string `json:"type" binding:"required"`
		Remark  string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "参数错误")
		return
	}

	dep := &models.Dependency{
		Name:    req.Name,
		Version: req.Version,
		Type:    req.Type,
		Remark:  req.Remark,
	}

	if err := c.service.Install(dep); err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	// 安装成功后保存到数据库
	c.service.Create(dep)

	utils.SuccessMsg(ctx, "安装成功")
}

// Uninstall 卸载依赖
func (c *DependencyController) Uninstall(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.BadRequest(ctx, "无效的 ID")
		return
	}

	// 获取依赖信息
	deps, _ := c.service.List("")
	var dep *models.Dependency
	for _, d := range deps {
		if d.ID == id {
			dep = &d
			break
		}
	}

	if dep == nil {
		utils.NotFound(ctx, "依赖不存在")
		return
	}

	if err := c.service.Uninstall(dep); err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	// 卸载成功后从数据库删除
	c.service.Delete(id)

	utils.SuccessMsg(ctx, "卸载成功")
}

// Reinstall 重新安装依赖
func (c *DependencyController) Reinstall(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.BadRequest(ctx, "无效的 ID")
		return
	}

	// 获取依赖信息
	deps, _ := c.service.List("")
	var dep *models.Dependency
	for _, d := range deps {
		if d.ID == id {
			dep = &d
			break
		}
	}

	if dep == nil {
		utils.NotFound(ctx, "依赖不存在")
		return
	}

	if err := c.service.Install(dep); err != nil {
		utils.ServerError(ctx, err.Error())
		return
	}

	utils.SuccessMsg(ctx, "重新安装成功")
}

// ReinstallAll 重新安装所有依赖
func (c *DependencyController) ReinstallAll(ctx *gin.Context) {
	depType := ctx.Query("type")
	if depType == "" {
		utils.BadRequest(ctx, "缺少 type 参数")
		return
	}

	deps, err := c.service.List(depType)
	if err != nil {
		utils.ServerError(ctx, "获取依赖列表失败")
		return
	}

	var failed []string
	for _, d := range deps {
		if err := c.service.Install(&d); err != nil {
			failed = append(failed, d.Name)
		}
	}

	if len(failed) > 0 {
		utils.ServerError(ctx, "部分包安装失败: "+strings.Join(failed, ", "))
		return
	}

	utils.SuccessMsg(ctx, "全部重新安装成功")
}

// GetInstalled 获取已安装的包
func (c *DependencyController) GetInstalled(ctx *gin.Context) {
	depType := ctx.Query("type")
	if depType == "" {
		utils.BadRequest(ctx, "缺少 type 参数")
		return
	}

	packages, err := c.service.GetInstalledPackages(depType)
	if err != nil {
		utils.ServerError(ctx, "获取已安装包失败: "+err.Error())
		return
	}

	utils.Success(ctx, packages)
}
