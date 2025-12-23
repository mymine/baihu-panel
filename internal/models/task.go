package models

import (
	"baihu/internal/constant"

	"gorm.io/gorm"
)

// CleanConfig 清理配置结构
type CleanConfig struct {
	Type string `json:"type"` // "day" 或 "count"
	Keep int    `json:"keep"` // 保留天数或条数
}

// Task represents a scheduled task
type Task struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Command     string         `json:"command" gorm:"type:text;not null"`
	Schedule    string         `json:"schedule" gorm:"size:100"`                // cron expression
	Timeout     int            `json:"timeout" gorm:"default:30"`               // 超时时间（分钟），默认30分钟
	WorkDir     string         `json:"work_dir" gorm:"size:255;default:''"`     // 工作目录，为空则使用 scripts 目录
	CleanConfig string         `json:"clean_config" gorm:"size:255;default:''"` // 清理配置 JSON
	Envs        string         `json:"envs" gorm:"size:255;default:''"`         // 环境变量ID列表，逗号分隔
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	LastRun     *LocalTime     `json:"last_run"`
	NextRun     *LocalTime     `json:"next_run"`
	CreatedAt   LocalTime      `json:"created_at"`
	UpdatedAt   LocalTime      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Task) TableName() string {
	return constant.TablePrefix + "tasks"
}

// TaskLog represents a log entry for task execution
type TaskLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"task_id" gorm:"index"`
	Command   string    `json:"command" gorm:"type:text"`
	Output    string    `json:"-" gorm:"type:longtext"` // gzip+base64 compressed
	Status    string    `json:"status" gorm:"size:20"`  // success, failed
	Duration  int64     `json:"duration"`               // milliseconds
	ExitCode  int       `json:"exit_code"`
	CreatedAt LocalTime `json:"created_at"`
}

func (TaskLog) TableName() string {
	return constant.TablePrefix + "task_logs"
}
