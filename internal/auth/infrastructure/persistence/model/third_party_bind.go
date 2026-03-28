package model

import "time"

// ThirdPartyBind 第三方登录绑定 GORM 模型
type ThirdPartyBind struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserId       uint64    `gorm:"column:user_id;not null;index"`
	Provider     int8      `gorm:"column:provider;type:tinyint;not null"`
	ProviderId   string    `gorm:"column:provider_id;type:varchar(100);not null"`
	ProviderName *string   `gorm:"column:provider_name;type:varchar(100)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// TableName 指定表名
func (ThirdPartyBind) TableName() string {
	return "third_party_binds"
}
