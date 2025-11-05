package models

// 회원 사용자 모델
type User struct {
	Username     string      `'json:"username"`
	PasswordHash string      `json:"-"`
	Profile      UserProfile `json:"profile"`
}

// LLM 세션용 사용자 프로필
type UserProfile struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}
