// Package service fdf
package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type (
	IAccountAdapterService interface {
		SendPasswordResetEmail(email string, baseURL string)
		SendConfirmationEmail(email string, baseURL string)
		Login(c *fiber.Ctx, user *types.User) error
		UserID(c *fiber.Ctx) (uint, error)
		UserClaims(c *fiber.Ctx) (*config.UserClaims, error)
		User(c *fiber.Ctx) (*types.User, error)
		IsLoggedIn(c *fiber.Ctx) bool
		Logout(c *fiber.Ctx) error
		AccessTokenCreate(c *fiber.Ctx, user *types.User) error
		RefreshTokenCreate(c *fiber.Ctx, user *types.User) error
		getAccountSession(duration time.Duration, c *fiber.Ctx) (*session.Session, error)
	}
	IAdapterService interface {
		SendPasswordResetEmail(email string, baseURL string, appCfg *config.AppCfg)
		SendConfirmationEmail(email string, baseURL string, appCfg *config.AppCfg)
		Login(c *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error
		UserID(c *fiber.Ctx, appCfg *config.AppCfg) (uint, error)
		UserClaims(c *fiber.Ctx, appCfg *config.AppCfg) (*config.UserClaims, error)
		User(c *fiber.Ctx, appCfg *config.AppCfg) (*types.User, error)
		IsLoggedIn(c *fiber.Ctx, appCfg *config.AppCfg) bool
		Logout(c *fiber.Ctx, appCfg *config.AppCfg) error
		AccessTokenCreate(c *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error
		RefreshTokenCreate(c *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error
		getAccountSession(duration time.Duration, c *fiber.Ctx, appCfg *config.AppCfg) (*session.Session, error)
	}
)

type accountAdapterService struct {
	adapterType IAdapterService
	*config.AppCfg
}

func (adapter *accountAdapterService) SendPasswordResetEmail(email string, baseURL string) {
	adapter.adapterType.SendPasswordResetEmail(email, baseURL, adapter.AppCfg)
}

func (adapter *accountAdapterService) SendConfirmationEmail(email string, baseURL string) {
	adapter.adapterType.SendConfirmationEmail(email, baseURL, adapter.AppCfg)
}

func (adapter *accountAdapterService) IsLoggedIn(c *fiber.Ctx) bool {
	return adapter.adapterType.IsLoggedIn(c, adapter.AppCfg)
}

func (adapter *accountAdapterService) Login(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.Login(c, user, adapter.AppCfg)
}

func (adapter *accountAdapterService) Logout(c *fiber.Ctx) error {
	return adapter.adapterType.Logout(c, adapter.AppCfg)
}

func (adapter *accountAdapterService) RefreshTokenCreate(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.RefreshTokenCreate(c, user, adapter.AppCfg)
}

func (adapter *accountAdapterService) AccessTokenCreate(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.AccessTokenCreate(c, user, adapter.AppCfg)
}

func (adapter *accountAdapterService) UserID(c *fiber.Ctx) (uint, error) {
	return adapter.adapterType.UserID(c, adapter.AppCfg)
}

func (adapter *accountAdapterService) UserClaims(c *fiber.Ctx) (*config.UserClaims, error) {
	return adapter.adapterType.UserClaims(c, adapter.AppCfg)
}

func (adapter *accountAdapterService) User(c *fiber.Ctx) (*types.User, error) {
	return adapter.adapterType.User(c, adapter.AppCfg)
}

func (adapter *accountAdapterService) getAccountSession(duration time.Duration, c *fiber.Ctx) (*session.Session, error) {
	return adapter.adapterType.getAccountSession(duration, c, adapter.AppCfg)
}

func NewAccountAdapterService(appCfg *config.AppCfg, adapterType IAdapterService) IAccountAdapterService {
	return &accountAdapterService{
		adapterType: adapterType,
		AppCfg:      appCfg,
	}
}

func GenerateConfirmURL(email string, baseURL string, appCfg *config.AppCfg) string {
	token, _ := pkg.Encrypt(email, appCfg.Server.Key, 15*time.Minute)
	uri := fmt.Sprintf("%s?t=%s", baseURL, token)
	return uri
}

func GenerateConfirmPath(context *fiber.Ctx, redirect string) string {
	if redirect == "" {
		redirect = fmt.Sprintf("%s/auth/verify-email", context.BaseURL())
	} else {
		redirect = fmt.Sprintf("%s/auth/verify-email", redirect)
	}
	return redirect
}

func GeneratePasswordResetURL(email string, baseURL string, appCfg *config.AppCfg) string {
	token, _ := pkg.Encrypt(email, appCfg.Server.Key, 15*time.Minute)
	uri := fmt.Sprintf("%s/?t=%s", baseURL, token)
	return uri
}

type AccountAdapterService struct {
	types.IUserService
}

func (adapter AccountAdapterService) SendPasswordResetEmail(email string, baseURL string, appCfg *config.AppCfg) {
	resetEmail := fmt.Sprintf("%s-reset-%d", email, time.Now().Unix())
	resetLink := GeneratePasswordResetURL(resetEmail, baseURL, appCfg)
	fmt.Println(resetLink)
	// htmlBody := app.Http.Mail.PrepareHtml("emails/password-reset", fiber.Map{
	// 	"reset_link": resetLink,
	// })
	// app.Http.Mail.Send(email, "You asked to reset? Please click here!", htmlBody, "", "")
}

func (adapter AccountAdapterService) SendConfirmationEmail(email string, baseURL string, appCfg *config.AppCfg) {
	confirmLink := GenerateConfirmURL(email, baseURL, appCfg)
	fmt.Println(confirmLink)
	// htmlBody := app.Http.Mail.PrepareHtml("emails/confirm", fiber.Map{
	// 	"confirm_link": confirmLink,
	// })
	// app.Http.Mail.Send(email, "Is it you? Please confirm!", htmlBody, "", "")
}

func (adapter AccountAdapterService) UserClaims(c *fiber.Ctx, appCfg *config.AppCfg) (*config.UserClaims, error) {
	fmt.Println("AccountAdapterService:UserClaims")
	accessToken := c.Cookies("access-token")
	fmt.Printf("AccountAdapterService::accessToken: %v", accessToken)
	accessClaim := &config.UserClaims{}
	if accessToken == "" {
		fmt.Println("Session setting.")
		sess, err := appCfg.Session.Get(c)
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
	accessClaim, err := appCfg.JwtSecrets.ParseAccessToken(accessToken)
	if err != nil {
		return accessClaim, err
	}
	return accessClaim, nil
}

func (adapter AccountAdapterService) User(c *fiber.Ctx, appCfg *config.AppCfg) (*types.User, error) {
	fmt.Println("AccountAdapterService:User")
	userClaims, err := adapter.UserClaims(c, appCfg)
	fmt.Printf("AccountAdapterService::userClaims: %v", userClaims)
	if err != nil {
		return &types.User{}, err
	}
	user := &types.User{}
	user, err = adapter.ViewByEmail(userClaims.Email, user)
	if err != nil {
		return &types.User{}, err
	}
	return user, nil
}

func (adapter AccountAdapterService) UserID(c *fiber.Ctx, appCfg *config.AppCfg) (uint, error) {
	accessClaim, err := adapter.UserClaims(c, appCfg)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(accessClaim.ID)
	if err != nil {
		return 1, err
	}
	return uint(id), nil
}

func (adapter AccountAdapterService) IsLoggedIn(context *fiber.Ctx, appCfg *config.AppCfg) bool {
	token := context.Cookies("refresh-token")
	if token == "" {
		if sess, err := appCfg.Session.Get(context); err != nil {
			return false
		} else {
			token = sess.Get("refresh-token").(string)
			if token != "" {
				parsedTokenClaim, err := appCfg.JwtSecrets.ParseRefreshToken(token)
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
		parsedTokenClaim, err := appCfg.JwtSecrets.ParseRefreshToken(token)
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

func (adapter AccountAdapterService) AccessTokenCreate(context *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error {
	accessClaims := config.NewUserClaims(
		*user,
		appCfg.JwtSecrets.AccessExpire,
	)
	accesstoken, err := appCfg.JwtSecrets.NewAccessToken(accessClaims)
	if err != nil {
		return err
	}
	expire, err := accessClaims.GetExpirationTime()

	if err == nil {
		if sess, err := adapter.getAccountSession(
			time.Until(expire.Time),
			context, appCfg); err != nil {
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

func (adapter AccountAdapterService) RefreshTokenCreate(context *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error {
	refreshClaims := config.NewUserClaims(
		*user,
		appCfg.JwtSecrets.RefreshExpire,
	)
	refreshtoken, err := appCfg.JwtSecrets.NewRefreshToken(refreshClaims)
	if err != nil {
		return err
	}
	expire, err := refreshClaims.GetExpirationTime()

	if err == nil {
		if sess, err := adapter.getAccountSession(time.Until(expire.Time), context, appCfg); err != nil {
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

func (adapter AccountAdapterService) getAccountSession(duration time.Duration, context *fiber.Ctx, appCfg *config.AppCfg) (*session.Session, error) {
	sess, errS := appCfg.Session.Get(context) // get/create new session
	sess.SetExpiry(duration)
	if errS != nil {
		return &session.Session{}, errS
	}

	return sess, nil
}

func (adapter AccountAdapterService) Login(context *fiber.Ctx, user *types.User, appCfg *config.AppCfg) error {
	if err := adapter.AccessTokenCreate(context, user, appCfg); err != nil {
		return err
	}

	if err := adapter.RefreshTokenCreate(context, user, appCfg); err != nil {
		return err
	}
	return nil
}

func (adapter AccountAdapterService) Logout(context *fiber.Ctx, appCfg *config.AppCfg) error {
	if sess, err := appCfg.Session.Get(context); err != nil {
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
