package auth

import (
	"crypto/subtle"
	"fmt"
	"strconv"
	"time"

	appUser "github.com/Prince-Letsyo/task-management-api-go/internal/user"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type (
	IUserClaims interface {
		Valid() error
		jwt.Claims
	}
	IJWT interface {
		newAccessToken(claims jwt.Claims) (string, error)
		newRefreshToken(claims jwt.Claims) (string, error)
		parseAccessToken(accessToken string) (*UserClaims, error)
		parseRefreshToken(refreshToken string) (*UserClaims, error)
		deleteToken(c *fiber.Ctx, token string) error
	}
)

type UserClaims struct {
	Id    string `json:"id"`
	First string `json:"first"`
	Last  string `json:"last"`
	jwt.RegisteredClaims
}

func (c UserClaims) Valid() error {
	vErr := new(ValidationError)
	now := time.Now().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if c.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt.Unix(), 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= ValidationErrorExpired
	}

	if c.VerifyIssuedAt(now, false) {
		vErr.Inner = fmt.Errorf("token used before issued")
		vErr.Errors |= ValidationErrorIssuedAt
	}

	if c.VerifyNotBefore(now, false) {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= ValidationErrorNotValidYet
	}

	if vErr.valid() {
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

// Compares the iss claim against cmp.
// If required is false, this method will return true if the value matches or is unset
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
func NewUserClaims(user appUser.User, expire int64) IUserClaims {
	return &UserClaims{
		Id:    strconv.Itoa(int(user.ID)),
		First: user.FirstName,
		Last:  user.LastName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
		},
	}
}
