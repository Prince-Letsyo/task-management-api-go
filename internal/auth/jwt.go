package auth

import (
	"errors"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ApiJWT struct {
	config.JwtSecrets
}

func (apijwt *ApiJWT) newAccessToken(claims jwt.Claims) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := accessToken.SignedString([]byte(apijwt.Api.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, nil
}

func (apijwt *ApiJWT) newRefreshToken(claims jwt.Claims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := refreshToken.SignedString([]byte(apijwt.Api.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, nil
}

func (apijwt *ApiJWT) parseAccessToken(accessToken string) (*UserClaims, error) {
	parsedAccessToken, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(apijwt.Api.Secret), nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedAccessToken.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("Invalid access token providered.")
}

func (apijwt *ApiJWT) parseRefreshToken(refreshToken string) (*UserClaims, error) {
	parsedRefreshToken, err := jwt.ParseWithClaims(
		refreshToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(apijwt.Api.Secret), nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedRefreshToken.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("Invalid refresh token providered.")
}

func (apijwt *ApiJWT) deleteToken(c *fiber.Ctx, token string) error {
	if _, err := apijwt.parseRefreshToken(token); err != nil {
		return err
	}
	c.ClearCookie(token)
	return nil
}

// NewUserClaims returns new UserClaims.
func NewJWT() IJWT {
	secret := app.Http.JwtSecrets
	return &ApiJWT{
		JwtSecrets: secret,
	}
}
