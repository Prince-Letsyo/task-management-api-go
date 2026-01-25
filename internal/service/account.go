// Package service fdf
package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

type AccountService struct {
	types.IUserService
	*config.AppCfg
}

func NewAccountService(appCfg *config.AppCfg, userService types.IUserService) *AccountService {
	return &AccountService{
		IUserService: userService,
		AppCfg:       appCfg,
	}
}

func (s *AccountService) SendPasswordResetEmail(email string, baseURL string) {
	resetEmail := fmt.Sprintf("%s-reset-%d", email, time.Now().Unix())
	resetLink, err := GeneratePasswordResetURL(resetEmail, baseURL, s.AppCfg)
	if err != nil {
		fmt.Printf("Error generating reset link: %v\n", err)
		return
	}

	htmlBody := s.AppCfg.Mail.PrepareHTML("emails/password-reset", fiber.Map{
		"Title":      "Password Reset",
		"reset_link": resetLink,
	})

	err = s.AppCfg.Mail.Send(
		email,
		"Password Reset Request",
		htmlBody,
		"",
		s.AppCfg.Mail.FromAddress,
	)
	if err != nil {
		fmt.Printf("Error sending password reset email: %v\n", err)
	}
}

func (s *AccountService) SendConfirmationEmail(email string, baseURL string) {
	confirmLink, err := GenerateConfirmURL(email, baseURL, s.AppCfg)
	if err != nil {
		fmt.Printf("Error generating confirm link: %v\n", err)
		return
	}

	htmlBody := s.AppCfg.Mail.PrepareHTML("emails/confirm", fiber.Map{
		"Title":        "Email Confirmation",
		"confirm_link": confirmLink,
	})

	err = s.AppCfg.Mail.Send(
		email,
		"Confirm Your Email",
		htmlBody,
		"",
		s.AppCfg.Mail.FromAddress,
	)
	if err != nil {
		fmt.Printf("Error sending confirmation email: %v\n", err)
	}
}

func (s *AccountService) UserClaims(c *fiber.Ctx) (*config.UserClaims, error) {
	fmt.Println("AccountAdapterService:UserClaims")
	accessToken := c.Cookies("access-token")
	fmt.Printf("AccountAdapterService::accessToken: %v", accessToken)
	accessClaim := &config.UserClaims{}
	if accessToken == "" {
		fmt.Println("Session setting.")
		sess, err := s.Session.Get(c)
		fmt.Println("Session got.")
		fmt.Printf("AccountAdapterService::Session err: %v ", err)
		if err != nil {
			return nil, err
		}
		if t := sess.Get("access-token"); t == nil {
			return accessClaim, errors.New("not loggedIn")
		} else {
			accessToken = t.(string)
			fmt.Printf("AccountAdapterService::Session accessToken: %v\n", accessToken)
		}
	}
	accessClaim, err := s.JwtSecrets.ParseAccessToken(accessToken)
	if err != nil {
		return accessClaim, err
	}
	return accessClaim, nil
}

func (s *AccountService) User(c *fiber.Ctx) (*types.User, error) {
	fmt.Println("AccountAdapterService:User")
	userClaims, err := s.UserClaims(c)
	fmt.Printf("AccountAdapterService::userClaims: %v", userClaims)
	if err != nil {
		return &types.User{}, err
	}
	user := &types.User{}
	user, err = s.ViewByEmail(userClaims.Email, user)
	if err != nil {
		return &types.User{}, err
	}
	return user, nil
}

func (s *AccountService) UserID(c *fiber.Ctx) (uint, error) {
	accessClaim, err := s.UserClaims(c)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(accessClaim.ID)
	if err != nil {
		return 1, err
	}
	return uint(id), nil
}

func (s *AccountService) IsLoggedIn(context *fiber.Ctx) bool {
	token := context.Cookies("refresh-token")
	if token == "" {
		if sess, err := s.Session.Get(context); err != nil {
			return false
		} else {
			token = sess.Get("refresh-token").(string)
			if token != "" {
				parsedTokenClaim, err := s.JwtSecrets.ParseRefreshToken(token)
				if err != nil {
					return false
				}
				if err = parsedTokenClaim.Valid(); err != nil {
					return false
				}
				token = parsedTokenClaim.ID
			} else {
				return false
			}
		}
	} else {
		parsedTokenClaim, err := s.JwtSecrets.ParseRefreshToken(token)
		if err != nil {
			return false
		}
		if err = parsedTokenClaim.Valid(); err != nil {
			return false
		}
		token = parsedTokenClaim.ID
	}
	return token != ""
}

func (s *AccountService) AccessTokenCreate(context *fiber.Ctx, user *types.User) error {
	accessClaims := config.NewUserClaims(
		*user,
		s.JwtSecrets.AccessExpire,
	)
	accesstoken, err := s.JwtSecrets.NewAccessToken(accessClaims)
	if err != nil {
		return err
	}
	expire, err := accessClaims.GetExpirationTime()

	if err == nil {
		if sess, err := s.getAccountSession(
			time.Until(expire.Time),
			context); err != nil {
			return err
		} else {
			sess.Set("access-token", accesstoken)
			if err = sess.Save(); err != nil {
				return err
			}
		}
		context.Cookie(&fiber.Cookie{
			Name:        "access-token",
			Value:       accesstoken,
			Secure:      false,
			HTTPOnly:    true,
			SessionOnly: true,
			Expires:     expire.Time,
		})
	} else {
		return err
	}
	return nil
}

func (s *AccountService) RefreshTokenCreate(context *fiber.Ctx, user *types.User) error {
	refreshClaims := config.NewUserClaims(
		*user,
		s.JwtSecrets.RefreshExpire,
	)
	refreshtoken, err := s.JwtSecrets.NewRefreshToken(refreshClaims)
	if err != nil {
		return err
	}
	expire, err := refreshClaims.GetExpirationTime()

	if err == nil {
		if sess, err := s.getAccountSession(time.Until(expire.Time), context); err != nil {
			return err
		} else {
			sess.Set("refresh-token", refreshtoken)
			if err = sess.Save(); err != nil {
				return err
			}
		}
		context.Cookie(&fiber.Cookie{
			Name:        "refresh-token",
			Value:       refreshtoken,
			Secure:      false,
			HTTPOnly:    true,
			SessionOnly: true,
			Expires:     expire.Time,
		})
	} else {
		return err
	}
	return nil
}

func (s *AccountService) getAccountSession(duration time.Duration, context *fiber.Ctx) (*session.Session, error) {
	sess, errS := s.Session.Get(context) // get/create new session
	sess.SetExpiry(duration)
	if errS != nil {
		return &session.Session{}, errS
	}

	return sess, nil
}

func (s *AccountService) Login(context *fiber.Ctx, user *types.User) error {
	if err := s.AccessTokenCreate(context, user); err != nil {
		return err
	}

	if err := s.RefreshTokenCreate(context, user); err != nil {
		return err
	}
	return nil
}

func (s *AccountService) Logout(context *fiber.Ctx) error {
	if sess, err := s.Session.Get(context); err != nil {
		return err
	} else {
		if err = sess.Destroy(); err != nil {
			return err
		}
	}
	context.ClearCookie()
	context.Set("X-DNS-Prefetch-Control", "off")
	context.Set("Pragma", "no-cache")
	context.Set("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
	context.Set("Cache-Control", "no-cache, must-revalidate, no-store, max-age=0, private")
	return nil
}

func GenerateConfirmURL(email string, baseURL string, appCfg *config.AppCfg) (string, error) {
	token, err := appCfg.Encryptor.Encrypt(email, 15*time.Minute)
	if err != nil {
		return "", err
	}
	uri := fmt.Sprintf("%s?t=%s", baseURL, token)
	return uri, nil
}

func GenerateConfirmPath(context *fiber.Ctx, redirect string) string {
	if redirect == "" {
		redirect = fmt.Sprintf("%s/auth/verify-email", context.BaseURL())
	} else {
		redirect = fmt.Sprintf("%s/auth/verify-email", redirect)
	}
	return redirect
}

func GeneratePasswordResetURL(email string, baseURL string, appCfg *config.AppCfg) (string, error) {
	token, err := appCfg.Encryptor.Encrypt(email, 15*time.Minute)
	if err != nil {
		return "", err
	}
	uri := fmt.Sprintf("%s/?t=%s", baseURL, token)
	return uri, nil
}
