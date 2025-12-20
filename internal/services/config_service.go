package services

import (
	"baihu/internal/constant"
	"encoding/json"
	"os"
)

type ServerConfig struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	SiteName string `json:"site_name"`
}

type DatabaseConfig struct {
	Type        string `json:"type"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	DBName      string `json:"dbname"`
	Path        string `json:"path"`
	TablePrefix string `json:"table_prefix"`
}

type SecurityConfig struct {
	JWTSecret    string `json:"jwt_secret"`
	PasswordSalt string `json:"password_salt"`
}

type TaskConfig struct {
	DefaultTimeout   int `json:"default_timeout"`
	LogRetentionDays int `json:"log_retention_days"`
}

type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Security SecurityConfig `json:"security"`
	Task     TaskConfig     `json:"task"`
}

var Config *AppConfig

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	Config = &AppConfig{}
	if err := json.Unmarshal(data, Config); err != nil {
		return nil, err
	}

	// 设置表前缀到 constant 包
	if Config.Database.TablePrefix != "" {
		constant.TablePrefix = Config.Database.TablePrefix
	}

	return Config, nil
}

func GetConfig() *AppConfig {
	return Config
}
