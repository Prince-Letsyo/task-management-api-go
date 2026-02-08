package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	Username       string `json:"username,omitempty"`
	Email          string `json:"email,omitempty"`
	UserID         uint   `json:"user_id,omitempty"`
	SID            string `json:"sid,omitempty"`
	RefreshVersion int    `json:"refresh_version,omitempty"`
	TokenType      string `json:"token_type,omitempty"`
	MFAPending     bool   `json:"mfa_pending,omitempty"`
	JTI            string `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

func buildClaims(base AuthClaims, expire time.Duration) AuthClaims {
	now := time.Now()
	base.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
	}
	return base
}

func createToken(secret string, claims AuthClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseToken(secret string, tokenString string) (*AuthClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsed.Claims.(*AuthClaims); ok {
		return claims, nil
	}
	return nil, jwt.ErrTokenMalformed
}

func hashToken(secret string, token string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func compareHash(a string, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}
