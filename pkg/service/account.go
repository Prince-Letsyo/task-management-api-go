package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
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
		getAAccountSession(duration time.Duration, c *fiber.Ctx) (*session.Session, error)
	}
)

type accountAdapterService struct {
	adapterType IAdapterService
}

func (adapter *accountAdapterService) SendPasswordResetEmail(email string, baseURL string) {
	adapter.adapterType.SendPasswordResetEmail(email, baseURL)
}

func (adapter *accountAdapterService) SendConfirmationEmail(email string, baseURL string) {
	adapter.adapterType.SendConfirmationEmail(email, baseURL)
}

func (adapter *accountAdapterService) IsLoggedIn(c *fiber.Ctx) bool {
	return adapter.adapterType.IsLoggedIn(c)
}

func (adapter *accountAdapterService) Login(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.Login(c, user)
}

func (adapter *accountAdapterService) Logout(c *fiber.Ctx) error {
	return adapter.adapterType.Logout(c)
}

func (adapter *accountAdapterService) RefreshTokenCreate(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.RefreshTokenCreate(c, user)
}

func (adapter *accountAdapterService) AccessTokenCreate(c *fiber.Ctx, user *types.User) error {
	return adapter.adapterType.AccessTokenCreate(c, user)
}

func (adapter *accountAdapterService) UserID(c *fiber.Ctx) (uint, error) {
	return adapter.adapterType.UserID(c)
}

func (adapter *accountAdapterService) UserClaims(c *fiber.Ctx) (*config.UserClaims, error) {
	return adapter.adapterType.UserClaims(c)
}

func (adapter *accountAdapterService) User(c *fiber.Ctx) (*types.User, error) {
	return adapter.adapterType.User(c)
}

func (adapter *accountAdapterService) getAccountSession(duration time.Duration, c *fiber.Ctx) (*session.Session, error) {
	return adapter.adapterType.getAAccountSession(duration, c)
}

func NewAccountAdapterService(adapterType IAdapterService) IAccountAdapterService {
	return &accountAdapterService{
		adapterType: adapterType,
	}
}

func GenerateConfirmURL(email string, baseURL string) string {
	token := pkg.Encrypt(email, app.Http.Server.Key)
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

func GeneratePasswordResetURL(email string, baseURL string) string {
	token := pkg.Encrypt(email, app.Http.Server.Key)
	uri := fmt.Sprintf("%s/?t=%s", baseURL, token)
	return uri
}

type AccountAdapterService struct {
	types.IUserService
	config.IJWT
}

func (adapter AccountAdapterService) SendPasswordResetEmail(email string, baseURL string) {
	resetEmail := fmt.Sprintf("%s-reset-%d", email, time.Now().Unix())
	resetLink := GeneratePasswordResetURL(resetEmail, baseURL)
	fmt.Println(resetLink)
	// htmlBody := app.Http.Mail.PrepareHtml("emails/password-reset", fiber.Map{
	// 	"reset_link": resetLink,
	// })
	// app.Http.Mail.Send(email, "You asked to reset? Please click here!", htmlBody, "", "")
}

func (adapter AccountAdapterService) SendConfirmationEmail(email string, baseURL string) {
	confirmLink := GenerateConfirmURL(email, baseURL)
	fmt.Println(confirmLink)
	// htmlBody := app.Http.Mail.PrepareHtml("emails/confirm", fiber.Map{
	// 	"confirm_link": confirmLink,
	// })
	// app.Http.Mail.Send(email, "Is it you? Please confirm!", htmlBody, "", "")
}

func (adapter AccountAdapterService) UserClaims(c *fiber.Ctx) (*config.UserClaims, error) {
	accessToken := c.Cookies("access-token")
	accessClaim, err := adapter.ParseAccessToken(accessToken)
	if err != nil {
		return &config.UserClaims{}, err
	}

	return accessClaim, nil
}

func (adapter AccountAdapterService) User(c *fiber.Ctx) (*types.User, error) {
	userClaims, err := adapter.UserClaims(c)
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

func (adapter AccountAdapterService) UserID(c *fiber.Ctx) (uint, error) {
	accessClaim, err := adapter.UserClaims(c)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(accessClaim.ID)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (adapter AccountAdapterService) IsLoggedIn(context *fiber.Ctx) bool {
	token := context.Cookies("access-token")
	return token != ""
}

func (adapter AccountAdapterService) AccessTokenCreate(context *fiber.Ctx, user *types.User) error {
	accessClaims := config.NewUserClaims(
		*user,
		adapter.GetAPIAccessExpireDuration(),
	)
	accesstoken, err := adapter.NewAccessToken(accessClaims)
	if err != nil {
		return err
	}
	expire, err := accessClaims.GetExpirationTime()

	if err == nil {
		if sess, err := adapter.getAccountSession(
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

func (adapter AccountAdapterService) RefreshTokenCreate(context *fiber.Ctx, user *types.User) error {
	refreshClaims := config.NewUserClaims(
		*user,
		adapter.GetAPIRefreshExpireDuration(),
	)
	refreshtoken, err := adapter.NewRefreshToken(refreshClaims)
	if err != nil {
		return err
	}
	expire, err := refreshClaims.GetExpirationTime()

	if err == nil {
		if sess, err := adapter.getAccountSession(time.Until(expire.Time), context); err != nil {
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

func (adapter AccountAdapterService) getAccountSession(duration time.Duration, context *fiber.Ctx) (*session.Session, error) {
	sess, errS := app.Http.Session.Get(context) // get/create new session
	sess.SetExpiry(duration)
	if errS != nil {
		return &session.Session{}, errS
	}

	return sess, nil
}

func (adapter AccountAdapterService) Login(context *fiber.Ctx, user *types.User) error {
	if err := adapter.AccessTokenCreate(context, user); err != nil {
		return err
	}

	if err := adapter.RefreshTokenCreate(context, user); err != nil {
		return err
	}
	return nil
}

func (adapter AccountAdapterService) Logout(context *fiber.Ctx) error {
	context.ClearCookie()
	context.Set("X-DNS-Prefetch-Control", "off")
	context.Set("Pragma", "no-cache")
	context.Set("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
	context.Set("Cache-Control", "no-cache, must-revalidate, no-store, max-age=0, private")
	return nil
}
