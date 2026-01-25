package config

import (
	"crypto/subtle"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oarkflow/log"
	"github.com/pkg/errors"
)

type (
	IUserClaims interface {
		Valid() error
		jwt.Claims
	}
)

type JwtSecrets struct {
	RefreshExpire time.Duration `yaml:"refresh"`
	AccessExpire  time.Duration `yaml:"access"`
	Secret        string        `yaml:"secret"`
}

func (tokenJWT *JwtSecrets) NewAccessToken(claims jwt.Claims) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := accessToken.SignedString([]byte(tokenJWT.Secret))
	if err != nil {
		return "", err
	}
	return signedStr, nil
}

func (tokenJWT *JwtSecrets) NewRefreshToken(claims jwt.Claims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, err := refreshToken.SignedString([]byte(tokenJWT.Secret))
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
			return []byte(tokenJWT.Secret), nil
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
			return []byte(tokenJWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedRefreshToken.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, errors.New("Invalid token providered.")
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

func Authenticate(appCfg *AppCfg, authConfig ...AuthConfig) fiber.Handler {
	// Init config
	var cfg AuthConfig
	if len(authConfig) > 0 {
		cfg = authConfig[0]
	}
	if cfg.SuccessHandler == nil {
		cfg.SuccessHandler = func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err error) error {
			var er fiber.Error
			if err.Error() == "missing or malformed jwt" {
				er.Code = fiber.StatusBadRequest
			} else {
				er.Code = fiber.StatusUnauthorized
			}
			er.Message = err.Error()

			return CustomErrorHandler(c, &er)
		}
	}
	if cfg.SigningKey == nil && len(cfg.SigningKeys) == 0 {
		log.Error().Msg("Fiber: JWT middleware requires signing key")
	}
	if cfg.SigningMethod == "" {
		cfg.SigningMethod = "HS256"
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = "user"
	}
	if cfg.Claims == nil {
		cfg.Claims = UserClaims{}
	}
	if cfg.TokenLookup == "" {
		cfg.TokenLookup = "header:" + fiber.HeaderAuthorization
	}
	if cfg.AuthScheme == "" {
		cfg.AuthScheme = "Bearer"
	}
	cfg.KeyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != cfg.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		if len(cfg.SigningKeys) > 0 {
			if kid, ok := t.Header["kid"].(string); ok {
				if key, ok := cfg.SigningKeys[kid]; ok {
					return key, nil
				}
			}
			return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
		}
		return cfg.SigningKey, nil
	}
	// Initialize
	extractors := make([]func(c *fiber.Ctx) (string, error), 0)
	rootParts := strings.Split(cfg.TokenLookup, ",")
	for _, rootPart := range rootParts {
		parts := strings.Split(strings.TrimSpace(rootPart), ":")

		switch parts[0] {
		case "header":
			extractors = append(extractors, jwtFromHeader(parts[1], cfg.AuthScheme))
		case "query":
			extractors = append(extractors, jwtFromQuery(parts[1]))
		case "param":
			extractors = append(extractors, jwtFromParam(parts[1]))
		case "cookie":
			extractors = append(extractors, jwtFromCookie(parts[1]))
		}
	}
	// Return middleware handler
	return func(c *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}
		var auth string
		var err error

		for _, extractor := range extractors {
			auth, err = extractor(c)
			if auth != "" && err == nil {
				break
			}
		}

		if err != nil {
			return cfg.ErrorHandler(c, err)
		}
		token := new(jwt.Token)
		if _, ok := cfg.Claims.(UserClaims); ok {
			token, err = jwt.Parse(auth, cfg.KeyFunc)
		} else {
			t := reflect.ValueOf(cfg.Claims).Type().Elem()
			claims := reflect.New(t).Interface().(jwt.Claims)
			token, err = jwt.ParseWithClaims(auth, claims, cfg.KeyFunc)
		}
		if err == nil && token.Valid {
			// Store user information from token into context.
			c.Locals(cfg.ContextKey, token)
			return cfg.SuccessHandler(c)
		}
		return cfg.ErrorHandler(c, err)
	}
}

// jwtFromHeader returns a function that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		auth := c.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", errors.New("missing or malformed jwt")
	}
}

// jwtFromQuery returns a function that extracts token from the query string.
func jwtFromQuery(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Query(param)
		if token == "" {
			return "", errors.New("missing or malformed jwt")
		}
		return token, nil
	}
}

// jwtFromParam returns a function that extracts token from the url param string.
func jwtFromParam(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Params(param)
		if token == "" {
			return "", errors.New("missing or malformed jwt")
		}
		return token, nil
	}
}

// jwtFromCookie returns a function that extracts token from the named cookie.
func jwtFromCookie(name string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Cookies(name)
		if token == "" {
			return "", errors.New("missing or malformed jwt")
		}
		return token, nil
	}
}

// Config defines the config for BasicAuth middleware
type AuthConfig struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid token.
	// Optional. Default: nil
	SuccessHandler fiber.Handler

	// ErrorHandler defines a function which is executed for an invalid token.
	// It may be used to define a custom JWT error.
	// Optional. Default: 401 Invalid or expired JWT
	ErrorHandler fiber.ErrorHandler

	// Signing key to validate token. Used as fallback if SigningKeys has length 0.
	// Required. This or SigningKeys.
	SigningKey interface{}

	// Map of signing keys to validate token with kid field usage.
	// Required. This or SigningKey.
	SigningKeys map[string]interface{}

	// Signing method, used to check token signing method.
	// Optional. Default: "HS256".
	// Possible values: "HS256", "HS384", "HS512", "ES256", "ES384", "ES512", "RS256", "RS384", "RS512"
	SigningMethod string

	// Context key to store user information from the token into context.
	// Optional. Default: "user".
	ContextKey string

	// Claims are extendable claims data defining token content.
	// Optional. Default value UserClaims
	Claims jwt.Claims

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "param:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// AuthScheme to be used in the Authorization header.
	// Optional. Default: "Bearer".
	AuthScheme string

	KeyFunc jwt.Keyfunc
}
