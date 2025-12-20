package services

import (
	"crypto/rand"
	"encoding/hex"

	"baihu/internal/constant"
	"baihu/internal/logger"
)

type InitService struct {
	settingsService *SettingsService
}

func NewInitService(settingsService *SettingsService) *InitService {
	return &InitService{
		settingsService: settingsService,
	}
}

// Initialize 执行系统初始化，返回 UserService
func (s *InitService) Initialize() *UserService {
	logger.Info("开始初始化系统...")

	// 初始化默认设置
	if err := s.settingsService.InitSettings(); err != nil {
		logger.Warnf("初始化设置失败: %v", err)
	}

	// 初始化 JWT Secret（也用作密码 salt，必须在创建 UserService 之前）
	s.initJWTSecret()

	// 创建 UserService（依赖 settingsService 获取 salt）
	userService := NewUserService(s.settingsService)
	// 创建管理员账号
	s.initializeAdmin(userService)

	return userService
}

// initializeAdmin 创建管理员账号
func (s *InitService) initializeAdmin(userService *UserService) {
	existingUser := userService.GetUserByUsername("admin")
	if existingUser != nil {
		logger.Info("管理员账号已存在，跳过创建")
		return
	}

	userService.CreateUser("admin", "123456", "admin@local", "admin")
	logger.Info("管理员账号创建成功: admin / 123456")
}

// IsInitialized 检查是否已初始化
func (s *InitService) IsInitialized() bool {
	return s.settingsService.Get(constant.SectionSystem, constant.KeyInitialized) == "true"
}

// initJWTSecret 初始化 JWT Secret，如果不存在则生成随机值
func (s *InitService) initJWTSecret() {
	existing := s.settingsService.Get(constant.SectionSystem, constant.KeyJWTSecret)
	if existing != "" {
		return
	}

	// 生成 32 字节随机密钥
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		logger.Warnf("生成 JWT Secret 失败: %v", err)
		return
	}

	secret := hex.EncodeToString(bytes)
	if err := s.settingsService.Set(constant.SectionSystem, constant.KeyJWTSecret, secret); err != nil {
		logger.Warnf("保存 JWT Secret 失败: %v", err)
		return
	}
	logger.Info("JWT Secret 已生成")
}
