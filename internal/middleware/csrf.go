package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CSRFToken(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			token := generateToken()
			c.SetCookie("csrf_token", token, 0, "/", "", secure, true)
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		cookieToken, _ := c.Cookie("csrf_token")
		formToken := c.PostForm("_csrf")
		headerToken := c.GetHeader("X-CSRF-Token")

		if cookieToken == "" || (formToken != cookieToken && headerToken != cookieToken) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
			return
		}

		c.Set("csrf_token", cookieToken)
		c.Next()
	}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
