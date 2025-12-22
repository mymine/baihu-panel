package deps_env

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"baihu/internal/constant"
	"baihu/internal/logger"
)

// getEnvsDir 获取虚拟环境存储目录
func getEnvsDir() string {
	return filepath.Join(constant.DataDir, "envs")
}

// CondaManager Conda 运行时管理器
type CondaManager struct {
	condaPath string
}

// NewCondaManager 创建 Conda 管理器
func NewCondaManager() *CondaManager {
	return &CondaManager{}
}

// GetType 获取运行时类型
func (cm *CondaManager) GetType() string {
	return "conda"
}

// IsAvailable 检查 Conda 是否可用
func (cm *CondaManager) IsAvailable() bool {
	path, err := cm.findCondaPath()
	if err != nil {
		return false
	}
	cm.condaPath = path
	return true
}

// findCondaPath 查找 conda 可执行文件路径
func (cm *CondaManager) findCondaPath() (string, error) {
	// 尝试常见的 conda 路径
	paths := []string{"conda", "micromamba", "/opt/conda/bin/conda", "/root/miniconda3/bin/conda", "/root/anaconda3/bin/conda"}
	for _, p := range paths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}
	return "", exec.ErrNotFound
}

// getCondaPath 获取 conda 路径
func (cm *CondaManager) getCondaPath() string {
	if cm.condaPath == "" {
		cm.findCondaPath()
	}
	return cm.condaPath
}

// condaEnvDetail 环境详情
type condaEnvDetail struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// condaEnvJSON conda env list --json 的输出结构
type condaEnvJSON struct {
	Envs        []string                  `json:"envs"`
	EnvsDetails map[string]condaEnvDetail `json:"envs_details"`
}

// ListEnvs 列出所有 Conda 环境
func (cm *CondaManager) ListEnvs() ([]RuntimeEnv, error) {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return nil, exec.ErrNotFound
	}

	cmd := exec.Command(condaPath, "env", "list", "--json")
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("Failed to list conda envs: %v", err)
		return nil, err
	}

	var envJSON condaEnvJSON
	if err := json.Unmarshal(output, &envJSON); err != nil {
		return nil, err
	}

	var envs []RuntimeEnv
	for _, envPath := range envJSON.Envs {
		detail, ok := envJSON.EnvsDetails[envPath]
		name := ""
		active := false
		if ok {
			name = detail.Name
			active = detail.Active
		}
		// 如果没有 name，使用路径
		if name == "" {
			name = envPath
		}
		envs = append(envs, RuntimeEnv{
			Name:   name,
			Path:   envPath,
			Active: active,
		})
	}

	return envs, nil
}

// CreateEnv 创建 Conda 环境
func (cm *CondaManager) CreateEnv(name string, version string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	envsDir := getEnvsDir()
	// 确保目录存在
	if err := os.MkdirAll(envsDir, 0755); err != nil {
		return err
	}

	envPath := filepath.Join(envsDir, name)
	args := []string{"create", "-p", envPath, "-y"}
	if version != "" {
		args = append(args, "python="+version)
	}

	logger.Infof("Creating conda env: %s %v", condaPath, args)
	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to create conda env: %v, output: %s", err, string(output))
		return err
	}
	logger.Infof("Conda env created: %s", envPath)

	// 写入 environments.txt
	if err := cm.appendToEnvironmentsTxt(envPath); err != nil {
		logger.Errorf("Failed to write environments.txt: %v", err)
	}

	return nil
}

// appendToEnvironmentsTxt 将环境路径追加到 environments.txt
func (cm *CondaManager) appendToEnvironmentsTxt(envPath string) error {
	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return err
	}

	envsDir := getEnvsDir()
	envsTxtPath := filepath.Join(envsDir, "environments.txt")

	// 读取现有内容，检查是否已存在
	content, _ := os.ReadFile(envsTxtPath)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == absPath {
			return nil // 已存在
		}
	}

	// 追加写入
	f, err := os.OpenFile(envsTxtPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(absPath + "\n")
	return err
}

// DeleteEnv 删除 Conda 环境
func (cm *CondaManager) DeleteEnv(name string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	if name == "base" {
		return nil // 不允许删除 base 环境
	}

	cmd := exec.Command(condaPath, "env", "remove", "-n", name, "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to delete conda env: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

// ListPackages 列出环境中的包
func (cm *CondaManager) ListPackages(envName string) ([]RuntimePackage, error) {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return nil, exec.ErrNotFound
	}

	args := []string{"list"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("Failed to list packages: %v", err)
		return nil, err
	}

	return parseCondaList(string(output)), nil
}

// parseCondaList 解析 conda list 输出
func parseCondaList(output string) []RuntimePackage {
	var packages []RuntimePackage
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		// 跳过注释和空行
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := fields[0]

			pkg := RuntimePackage{
				Name:    name,
				Version: fields[1],
			}
			if len(fields) >= 4 {
				pkg.Channel = fields[3]
			}
			packages = append(packages, pkg)
		}
	}

	return packages
}

// InstallPackage 安装包
func (cm *CondaManager) InstallPackage(envName string, packageName string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	args := []string{"install", "-y"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}
	args = append(args, packageName)

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to install package: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

// UninstallPackage 卸载包
func (cm *CondaManager) UninstallPackage(envName string, packageName string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	args := []string{"remove", "-y"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}
	args = append(args, packageName)

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to uninstall package: %v, output: %s", err, string(output))
		return err
	}

	return nil
}
