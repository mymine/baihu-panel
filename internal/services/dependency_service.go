package services

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/logger"
	"github.com/engigu/baihu-panel/internal/models"
)

type DependencyService struct{}

func NewDependencyService() *DependencyService {
	return &DependencyService{}
}

// List 获取依赖列表
func (s *DependencyService) List(depType string) ([]models.Dependency, error) {
	var deps []models.Dependency
	query := database.DB
	if depType != "" {
		query = query.Where("type = ?", depType)
	}
	err := query.Order("id desc").Find(&deps).Error
	return deps, err
}

// Create 创建依赖记录
func (s *DependencyService) Create(dep *models.Dependency) error {
	// 检查是否已存在
	var existing models.Dependency
	if err := database.DB.Where("name = ? AND type = ?", dep.Name, dep.Type).First(&existing).Error; err == nil {
		return errors.New("依赖已存在")
	}
	return database.DB.Create(dep).Error
}

// Delete 删除依赖记录
func (s *DependencyService) Delete(id int) error {
	return database.DB.Delete(&models.Dependency{}, id).Error
}

// Install 安装依赖
func (s *DependencyService) Install(dep *models.Dependency) error {
	var cmd *exec.Cmd
	var packageSpec string

	if dep.Version != "" {
		if dep.Type == "py" {
			packageSpec = dep.Name + "==" + dep.Version
		} else {
			packageSpec = dep.Name + "@" + dep.Version
		}
	} else {
		packageSpec = dep.Name
	}

	switch dep.Type {
	case "py":
		cmd = exec.Command("pip", "install", packageSpec)
	case "node":
		cmd = exec.Command("npm", "install", "-g", packageSpec)
	default:
		return errors.New("不支持的依赖类型")
	}

	logger.Infof("Installing %s package: %s", dep.Type, packageSpec)
	output, err := cmd.CombinedOutput()
	dep.Log = string(output)

	if err != nil {
		logger.Errorf("Install failed: %v, output: %s", err, string(output))
		return errors.New("安装失败: " + string(output))
	}
	logger.Infof("Install success: %s", packageSpec)

	return nil
}

// Uninstall 卸载依赖
func (s *DependencyService) Uninstall(dep *models.Dependency) error {
	var cmd *exec.Cmd

	switch dep.Type {
	case "py":
		cmd = exec.Command("pip", "uninstall", "-y", dep.Name)
	case "node":
		cmd = exec.Command("npm", "uninstall", "-g", dep.Name)
	default:
		return errors.New("不支持的依赖类型")
	}

	logger.Infof("Uninstalling %s package: %s", dep.Type, dep.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Uninstall failed: %v, output: %s", err, string(output))
		return errors.New("卸载失败: " + string(output))
	}

	return nil
}

// GetInstalledPackages 获取已安装的包列表
func (s *DependencyService) GetInstalledPackages(depType string) ([]models.Dependency, error) {
	var packages []models.Dependency

	switch depType {
	case "py":
		return s.getPipPackages()
	case "node":
		return s.getNpmPackages()
	default:
		return packages, errors.New("不支持的依赖类型")
	}
}

// getPipPackages 获取 pip 已安装的包
func (s *DependencyService) getPipPackages() ([]models.Dependency, error) {
	cmd := exec.Command("pip", "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []models.Dependency
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "==", 2)
		pkg := models.Dependency{
			Name: parts[0],
			Type: "py",
		}
		if len(parts) > 1 {
			pkg.Version = parts[1]
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

// getNpmPackages 获取 npm 全局安装的包
func (s *DependencyService) getNpmPackages() ([]models.Dependency, error) {
	cmd := exec.Command("npm", "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		// npm list 在没有包时也会返回错误，忽略
	}

	var packages []models.Dependency
	// 简单解析，不用 json 库
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"version"`) {
			continue
		}
		if strings.HasPrefix(line, `"`) && strings.Contains(line, ":") {
			// 格式: "package-name": {
			name := strings.Trim(strings.Split(line, ":")[0], `" `)
			if name != "" && name != "dependencies" {
				packages = append(packages, models.Dependency{
					Name: name,
					Type: "node",
				})
			}
		}
	}

	return packages, nil
}
