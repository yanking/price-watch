package mysql

import "time"

// ThirdPartyBindPO 第三方登录绑定持久化对象
type ThirdPartyBindPO struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserId       int64     `gorm:"column:user_id;not null;index"`
	Provider     int8      `gorm:"column:provider;type:tinyint;not null"`
	ProviderId   string    `gorm:"column:provider_id;type:varchar(100);not null"`
	ProviderName *string   `gorm:"column:provider_name;type:varchar(100)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// TableName 指定表名
func (ThirdPartyBindPO) TableName() string {
	return "third_party_binds"
}
