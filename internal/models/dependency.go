package models

import (
	"github.com/engigu/baihu-panel/internal/constant"
	"time"
)

// Dependency 依赖包模型
type Dependency struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Version   string    `json:"version" gorm:"size:50"`
	Type      string    `json:"type" gorm:"size:10;not null"` // py 或 node
	Remark    string    `json:"remark" gorm:"size:255"`
	Log       string    `json:"log" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Dependency) TableName() string {
	return constant.TablePrefix + "deps"
}
