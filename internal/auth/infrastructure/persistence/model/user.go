package model

import "time"

// User 用户 GORM 模型
type User struct {
	Id            uint64    `gorm:"column:id;primaryKey"`
	Username      string    `gorm:"column:username;type:varchar(50);uniqueIndex"`
	PasswordHash  *string   `gorm:"column:password_hash;type:varchar(255)"`
	Email         *string   `gorm:"column:email;type:varchar(100)"`
	EmailVerified bool      `gorm:"column:email_verified"`
	AreaCode      *string   `gorm:"column:area_code;type:varchar(10)"`
	Phone         *string   `gorm:"column:phone;type:varchar(20)"`
	PhoneVerified bool      `gorm:"column:phone_verified"`
	Avatar        *string   `gorm:"column:avatar;type:varchar(500)"`
	Nickname      *string   `gorm:"column:nickname;type:varchar(50)"`
	Status        int8      `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// statusToInt8 将 UserStatus string 转换为 int8
func StatusToInt8(status string) int8 {
	switch status {
	case "active":
		return 1
	case "inactive":
		return 2
	case "suspended":
		return 3
	case "deleted":
		return 4
	default:
		return 1
	}
}

// int8ToStatus 将 int8 转换为 UserStatus string
func Int8ToStatus(status int8) string {
	switch status {
	case 1:
		return "active"
	case 2:
		return "inactive"
	case 3:
		return "suspended"
	case 4:
		return "deleted"
	default:
		return "active"
	}
}
