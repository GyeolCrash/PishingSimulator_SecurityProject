/* JWT 토큰 생성 및 검증을 위한 유틸리티 함수들 */

package auth

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey []byte

// JWT 키 초기화, 런타임에 자동 호출
func init() {
	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	log.Printf("JWT Key: %s", jwtKey)
	if len(jwtKey) == 0 {
		jwtKey = []byte("default_secret_key") // 기본 키 설정 (권장하지 않음)
		log.Println("Warning: JWT_SECRET_KEY environment variable is not set. Using default key.")
	}
}

// Claims 구조체 정의, JWT 페이로드에 사용자명 포함
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT 토큰 생성
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "PishingSimulator-api",
			Subject:   "user_auth_token",
		},
	}

	// 토큰 문자열 생성 및 서명
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// JWT 토큰 검증
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	// 토큰 파싱 및 검증
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	// 만약을 위한 토큰 유효성 재검사
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
