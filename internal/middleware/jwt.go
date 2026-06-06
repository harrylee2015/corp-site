package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTMiddleware struct {
	secret []byte
	secure bool
}

func NewJWTMiddleware(secret string, secure bool) *JWTMiddleware {
	return &JWTMiddleware{secret: []byte(secret), secure: secure}
}

func (j *JWTMiddleware) GenerateToken(userID, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTMiddleware) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (j *JWTMiddleware) SetTokenCookies(c *gin.Context, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	c.SetCookie("access_token", accessToken, int(accessTTL.Seconds()), "/", "", j.secure, true)
	if refreshToken != "" {
		c.SetCookie("refresh_token", refreshToken, int(refreshTTL.Seconds()), "/", "", j.secure, true)
	}
}

func (j *JWTMiddleware) ClearTokenCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", j.secure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", j.secure, true)
}

func (j *JWTMiddleware) AuthRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		if cookie, err := c.Cookie("access_token"); err == nil {
			tokenStr = cookie
		}
		if tokenStr == "" {
			header := c.GetHeader("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				tokenStr = strings.TrimPrefix(header, "Bearer ")
			}
		}
		if tokenStr == "" {
			redirectTo := j.loginRedirect(allowedRoles)
			c.Redirect(http.StatusFound, redirectTo)
			c.Abort()
			return
		}

		claims, err := j.ParseToken(tokenStr)
		if err != nil {
			refreshCookie, refreshErr := c.Cookie("refresh_token")
			if refreshErr == nil {
				refreshClaims, parseErr := j.ParseToken(refreshCookie)
				if parseErr == nil {
					newAccess, genErr := j.GenerateToken(refreshClaims.UserID, refreshClaims.Role, 2*time.Hour)
					if genErr != nil {
						j.ClearTokenCookies(c)
						c.Redirect(http.StatusFound, j.loginRedirect(allowedRoles))
						c.Abort()
						return
					}
					j.SetTokenCookies(c, newAccess, refreshCookie, 2*time.Hour, 0)
					c.Set("user_id", refreshClaims.UserID)
					c.Set("role", refreshClaims.Role)
					c.Next()
					return
				}
			}
			j.ClearTokenCookies(c)
			c.Redirect(http.StatusFound, j.loginRedirect(allowedRoles))
			c.Abort()
			return
		}

		if len(allowedRoles) > 0 {
			allowed := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (j *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		if cookie, err := c.Cookie("access_token"); err == nil {
			tokenStr = cookie
		}
		if tokenStr == "" {
			header := c.GetHeader("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				tokenStr = strings.TrimPrefix(header, "Bearer ")
			}
		}
		if tokenStr == "" {
			c.Next()
			return
		}

		claims, err := j.ParseToken(tokenStr)
		if err != nil {
			refreshCookie, refreshErr := c.Cookie("refresh_token")
			if refreshErr == nil {
				refreshClaims, parseErr := j.ParseToken(refreshCookie)
				if parseErr == nil {
					newAccess, genErr := j.GenerateToken(refreshClaims.UserID, refreshClaims.Role, 2*time.Hour)
					if genErr == nil {
						j.SetTokenCookies(c, newAccess, refreshCookie, 2*time.Hour, 0)
						c.Set("user_id", refreshClaims.UserID)
						c.Set("role", refreshClaims.Role)
						c.Next()
						return
					}
				}
			}
			j.ClearTokenCookies(c)
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (j *JWTMiddleware) loginRedirect(allowedRoles []string) string {
	for _, r := range allowedRoles {
		if r == "admin" {
			return "/admin/login"
		}
	}
	return "/login"
}
