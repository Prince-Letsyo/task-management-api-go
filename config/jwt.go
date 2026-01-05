package config

import (
	"crypto/subtle"
	"fmt"
	"strconv"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type (
	IUserClaims interface {
		Valid() error
		jwt.Claims
	}
)

type JwtConfig struct {
	Secret string `yaml:"secret"`
	Expire int64  `yaml:"expire"`
}

type JwtSecrets struct {
	RefreshExpire int64     `yaml:"refresh_expire"`
	AccessExpire  int64     `yaml:"access_expire"`
	App           JwtConfig `yaml:"app"`
	API           JwtConfig `yaml:"api"`
}

func (tokenJWT *JwtSecrets) NewAccessToken(claims jwt.Claims) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := accessToken.SignedString([]byte(tokenJWT.API.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, nil
}

func (tokenJWT *JwtSecrets) NewRefreshToken(claims jwt.Claims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := refreshToken.SignedString([]byte(tokenJWT.API.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, nil
}

func (tokenJWT *JwtSecrets) ParseAccessToken(accessToken string) (*UserClaims, error) {
	parsedAccessToken, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenJWT.API.Secret), nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedAccessToken.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("Invalid token providered.")
}

func (tokenJWT *JwtSecrets) ParseRefreshToken(refreshToken string) (*UserClaims, error) {
	parsedRefreshToken, err := jwt.ParseWithClaims(
		refreshToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenJWT.API.Secret), nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedRefreshToken.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("Invalid token providered.")
}

func (tokenJWT *JwtSecrets) GetAPPConfig() *JwtConfig {
	return &tokenJWT.App
}

func (tokenJWT *JwtSecrets) GetAPIConfig() *JwtConfig {
	return &tokenJWT.API
}

func (tokenJWT *JwtSecrets) GetAPIAccessExpireDuration() time.Duration {
	return time.Duration(tokenJWT.AccessExpire) * time.Minute
}

func (tokenJWT *JwtSecrets) GetAPIRefreshExpireDuration() time.Duration {
	return time.Duration(tokenJWT.RefreshExpire) * time.Hour
}

type UserClaims struct {
	ID    string `json:"id"`
	First string `json:"first"`
	Last  string `json:"last"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (c UserClaims) Valid() error {
	vErr := new(pkg.ValidationError)
	now := time.Now().Unix()

	// The  below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if c.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt.Unix(), 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= pkg.ValidationErrorExpired
	}

	if c.VerifyIssuedAt(now, false) {
		vErr.Inner = fmt.Errorf("token used before issued")
		vErr.Errors |= pkg.ValidationErrorIssuedAt
	}

	if c.VerifyNotBefore(now, false) {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= pkg.ValidationErrorNotValidYet
	}

	if vErr.Valid() {
		return nil
	}

	return vErr
}

// Compares the aud claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c UserClaims) VerifyAudience(cmp string, req bool) bool {
	return verifyAud(c.Audience, cmp, req)
}

// Compares the exp claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c UserClaims) VerifyExpiresAt(cmp int64, req bool) bool {
	return verifyExp(c.ExpiresAt.Unix(), cmp, req)
}

// Compares the iat claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c UserClaims) VerifyIssuedAt(cmp int64, req bool) bool {
	return verifyIat(c.IssuedAt.Unix(), cmp, req)
}

// claim against cmp.
// If rfalse, this method will return true if the value matches or is unset
func (c UserClaims) VerifyIssuer(cmp string, req bool) bool {
	return verifyIss(c.Issuer, cmp, req)
}

// Compares the nbf claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c UserClaims) VerifyNotBefore(cmp int64, req bool) bool {
	return verifyNbf(c.NotBefore.Unix(), cmp, req)
}

// ----- helpers

func verifyAud(aud []string, cmp string, required bool) bool {
	if len(aud) == 0 {
		return !required
	}

	for _, a := range aud {
		if subtle.ConstantTimeCompare([]byte(a), []byte(cmp)) != 0 {
			return true
		}
	}
	return false
}

func verifyExp(exp int64, now int64, required bool) bool {
	if exp == 0 {
		return !required
	}
	return now <= exp
}

func verifyIat(iat int64, now int64, required bool) bool {
	if iat == 0 {
		return !required
	}
	return now >= iat
}

func verifyIss(iss string, cmp string, required bool) bool {
	if iss == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(iss), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func verifyNbf(nbf int64, now int64, required bool) bool {
	if nbf == 0 {
		return !required
	}
	return now >= nbf
}

// NewUserClaims returns new UserClaims.
func NewUserClaims(user types.User, expire time.Duration) IUserClaims {
	return &UserClaims{
		ID:    strconv.Itoa(int(user.ID)),
		First: user.FirstName,
		Last:  user.LastName,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
	}
}
