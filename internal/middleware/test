package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"BackendFramework/internal/config"
	"BackendFramework/internal/database"
	"BackendFramework/internal/model"
)

// ✅ Ganti variable global dengan function
func getJwtSecret() []byte {
	return []byte(config.JWT_SIGNATURE_KEY)
}

type AccessClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID string) (string, error) {
	claims := &AccessClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "BackendFramework UIB",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJwtSecret()) // ✅
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	plainText := base64.StdEncoding.EncodeToString(b)

	key := []byte(config.ENCRYPTION_KEY)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func ValidateToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getJwtSecret(), nil // ✅
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	var accessToken model.TokenData
	err = database.DbWebkita.
		Where("user_id = ?", claims.UserID).
		First(&accessToken).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("token not found or expired")
	}
	if err != nil {
		return nil, errors.New("failed to validate token")
	}
	if accessToken.AccessToken != tokenString || strings.ToLower(accessToken.IsValidToken) != "y" {
		return nil, errors.New("token not found or expired")
	}

	return claims, nil
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Authorization token not provided",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}