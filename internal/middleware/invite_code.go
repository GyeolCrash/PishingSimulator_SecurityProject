package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func InviteCodeMiddleware() gin.HandlerFunc {
	inviteCode := os.Getenv("SIGNUP_INVIT_CODE")
	if inviteCode == "" {
		log.Fatal("[Fatal] SIGNUP_INVITE_CODE가 없음, 회원가입 불가")
	}
	return func(c *gin.Context) {
		clientKey := c.GetHeader("X-Invite-Code")

		if clientKey != inviteCode {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invaid"})
			return
		}
		c.Next()
	}
}
