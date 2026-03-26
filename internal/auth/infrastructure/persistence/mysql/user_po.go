package mysql

import "time"

// UserPO 用户持久化对象
type UserPO struct {
	Id            int64     `gorm:"column:id;primaryKey"`
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
func (UserPO) TableName() string {
	return "users"
}
