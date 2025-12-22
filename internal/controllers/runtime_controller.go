package controllers

import (
	"baihu/internal/services/deps_env"
	"baihu/internal/utils"

	"github.com/gin-gonic/gin"
)

type RuntimeController struct {
	runtimeService *deps_env.RuntimeService
}

func NewRuntimeController() *RuntimeController {
	return &RuntimeController{
		runtimeService: deps_env.NewRuntimeService(),
	}
}

// GetAvailableRuntimes 获取可用的运行时列表
func (rc *RuntimeController) GetAvailableRuntimes(c *gin.Context) {
	runtimes := rc.runtimeService.GetAvailableRuntimes()
	utils.Success(c, runtimes)
}

// ListEnvs 列出指定运行时的所有环境
func (rc *RuntimeController) ListEnvs(c *gin.Context) {
	runtimeType := c.Query("type")
	if runtimeType == "" {
		utils.BadRequest(c, "缺少 type 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	if !manager.IsAvailable() {
		utils.BadRequest(c, "运行时不可用")
		return
	}

	envs, err := manager.ListEnvs()
	if err != nil {
		utils.ServerError(c, "获取环境列表失败: "+err.Error())
		return
	}

	utils.Success(c, envs)
}

// CreateEnv 创建环境
func (rc *RuntimeController) CreateEnv(c *gin.Context) {
	runtimeType := c.Query("type")
	if runtimeType == "" {
		utils.BadRequest(c, "缺少 type 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Version string `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := manager.CreateEnv(req.Name, req.Version); err != nil {
		utils.ServerError(c, "创建环境失败: "+err.Error())
		return
	}

	utils.SuccessMsg(c, "环境创建成功")
}

// DeleteEnv 删除环境
func (rc *RuntimeController) DeleteEnv(c *gin.Context) {
	runtimeType := c.Query("type")
	envName := c.Query("name")

	if runtimeType == "" || envName == "" {
		utils.BadRequest(c, "缺少 type 或 name 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	if envName == "base" {
		utils.BadRequest(c, "不能删除 base 环境")
		return
	}

	if err := manager.DeleteEnv(envName); err != nil {
		utils.ServerError(c, "删除环境失败: "+err.Error())
		return
	}

	utils.SuccessMsg(c, "环境删除成功")
}

// ListPackages 列出环境中的包
func (rc *RuntimeController) ListPackages(c *gin.Context) {
	runtimeType := c.Query("type")
	envName := c.Query("env")

	if runtimeType == "" || envName == "" {
		utils.BadRequest(c, "缺少 type 或 env 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	packages, err := manager.ListPackages(envName)
	if err != nil {
		utils.ServerError(c, "获取包列表失败: "+err.Error())
		return
	}

	utils.Success(c, packages)
}

// InstallPackage 安装包
func (rc *RuntimeController) InstallPackage(c *gin.Context) {
	runtimeType := c.Query("type")
	envName := c.Query("env")

	if runtimeType == "" || envName == "" {
		utils.BadRequest(c, "缺少 type 或 env 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	var req struct {
		Package string `json:"package" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := manager.InstallPackage(envName, req.Package); err != nil {
		utils.ServerError(c, "安装包失败: "+err.Error())
		return
	}

	utils.SuccessMsg(c, "包安装成功")
}

// UninstallPackage 卸载包
func (rc *RuntimeController) UninstallPackage(c *gin.Context) {
	runtimeType := c.Query("type")
	envName := c.Query("env")

	if runtimeType == "" || envName == "" {
		utils.BadRequest(c, "缺少 type 或 env 参数")
		return
	}

	manager := rc.runtimeService.GetManager(runtimeType)
	if manager == nil {
		utils.NotFound(c, "运行时类型不存在")
		return
	}

	var req struct {
		Package string `json:"package" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := manager.UninstallPackage(envName, req.Package); err != nil {
		utils.ServerError(c, "卸载包失败: "+err.Error())
		return
	}

	utils.SuccessMsg(c, "包卸载成功")
}
