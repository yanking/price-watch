package valueobject

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	hash string
}

func NewPassword(plain string) (*Password, error) {
	if plain == "" {
		return nil, nil
	}

	// 验证长度
	if len(plain) < 8 {
		return nil, errors.New("密码长度不能少于8位")
	}
	if len(plain) > 20 {
		return nil, errors.New("密码长度不能超过20位")
	}

	// 验证必须包含字母
	hasLetter, err := regexp.MatchString(`[a-zA-Z]`, plain)
	if err != nil {
		return nil, errors.New("正则表达式匹配失败")
	}
	if !hasLetter {
		return nil, errors.New("密码必须包含字母")
	}

	// 验证必须包含数字
	hasDigit, err := regexp.MatchString(`[0-9]`, plain)
	if err != nil {
		return nil, errors.New("正则表达式匹配失败")
	}
	if !hasDigit {
		return nil, errors.New("密码必须包含数字")
	}

	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	return &Password{hash: string(hash)}, nil
}

func NewPasswordFromHash(hash string) *Password {
	return &Password{hash: hash}
}

func (p *Password) Hash() string {
	return p.hash
}

func (p *Password) Verify(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plain))
	return err == nil
}
