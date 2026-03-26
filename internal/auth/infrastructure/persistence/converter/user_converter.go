package converter

import (
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/mysql"
)

// UserToPO 将领域对象 User 转换为持久化对象 UserPO
func UserToPO(user *entity.User) *mysql.UserPO {
	po := &mysql.UserPO{
		Id:            int64(user.ID()),
		Username:      user.Username(),
		EmailVerified: user.EmailVerified(),
		PhoneVerified: user.PhoneVerified(),
		Status:        userStatusToInt8(user.Status()),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}

	// 处理密码哈希
	if user.Password() != nil {
		hash := user.Password().Hash()
		po.PasswordHash = &hash
	}

	// 处理邮箱
	if user.Email() != nil {
		email := user.Email().Value()
		po.Email = &email
	}

	// 处理区号
	if user.AreaCode() != "" {
		areaCode := user.AreaCode()
		po.AreaCode = &areaCode
	}

	// 处理手机号
	if user.Phone() != "" {
		phone := user.Phone()
		po.Phone = &phone
	}

	return po
}

// POToUser 将持久化对象 UserPO 转换为领域对象 User
func POToUser(po *mysql.UserPO) *entity.User {
	// 构建密码哈希字符串
	var passwordHash string
	if po.PasswordHash != nil {
		passwordHash = *po.PasswordHash
	}

	// 构建邮箱字符串
	var emailStr string
	if po.Email != nil {
		emailStr = *po.Email
	}

	// 构建区号字符串
	var areaCodeStr string
	if po.AreaCode != nil {
		areaCodeStr = *po.AreaCode
	}

	// 构建手机号字符串
	var phoneStr string
	if po.Phone != nil {
		phoneStr = *po.Phone
	}

	user, err := entity.NewUserFromData(
		uint64(po.Id),
		po.Username,
		passwordHash,
		emailStr,
		areaCodeStr,
		phoneStr,
		po.EmailVerified,
		po.PhoneVerified,
		string(int8ToUserStatus(po.Status)),
		"", // oauthProvider - 暂不处理
		"", // oauthID - 暂不处理
		po.CreatedAt,
		po.UpdatedAt,
	)
	if err != nil {
		// 理论上不应该发生，因为数据来自数据库
		// 如果发生，返回一个基本的用户对象
		user, _ = entity.NewUserFromData(
			uint64(po.Id),
			po.Username,
			"",
			"",
			"",
			"",
			po.EmailVerified,
			po.PhoneVerified,
			"active",
			"",
			"",
			po.CreatedAt,
			po.UpdatedAt,
		)
	}

	return user
}

// ThirdPartyBindToPO 将领域对象 ThirdPartyBind 转换为持久化对象 ThirdPartyBindPO
func ThirdPartyBindToPO(bind *entity.ThirdPartyBind) *mysql.ThirdPartyBindPO {
	po := &mysql.ThirdPartyBindPO{
		Id:         bind.Id(),
		UserId:     bind.UserId(),
		Provider:   int8(bind.Provider()),
		ProviderId: bind.ProviderId(),
		CreatedAt:  bind.CreatedAt(),
	}

	if bind.ProviderName() != "" {
		providerName := bind.ProviderName()
		po.ProviderName = &providerName
	}

	return po
}

// POToThirdPartyBind 将持久化对象 ThirdPartyBindPO 转换为领域对象 ThirdPartyBind
func POToThirdPartyBind(po *mysql.ThirdPartyBindPO) *entity.ThirdPartyBind {
	bind := entity.NewThirdPartyBind(
		po.UserId,
		valueobject.OAuthProvider(po.Provider),
		po.ProviderId,
		"",
	)

	// 设置 ID
	bind.SetId(po.Id)

	// 设置 ProviderName
	if po.ProviderName != nil {
		bind.SetProviderName(*po.ProviderName)
	}

	// 设置 CreatedAt
	bind.SetCreatedAt(po.CreatedAt)

	return bind
}

// userStatusToInt8 将 UserStatus 转换为 int8
func userStatusToInt8(status entity.UserStatus) int8 {
	switch status {
	case entity.UserStatusActive:
		return 1
	case entity.UserStatusInactive:
		return 2
	case entity.UserStatusSuspended:
		return 3
	case entity.UserStatusDeleted:
		return 4
	default:
		return 1 // 默认为激活状态
	}
}

// int8ToUserStatus 将 int8 转换为 UserStatus
func int8ToUserStatus(status int8) entity.UserStatus {
	switch status {
	case 1:
		return entity.UserStatusActive
	case 2:
		return entity.UserStatusInactive
	case 3:
		return entity.UserStatusSuspended
	case 4:
		return entity.UserStatusDeleted
	default:
		return entity.UserStatusActive
	}
}
