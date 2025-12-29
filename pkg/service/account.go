package service

import (
	"errors"
	"fmt"
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
		User(c *fiber.Ctx) (*types.User, error)
		UserID(c *fiber.Ctx) (uint, error)
		Login(c *fiber.Ctx, user *types.User) error
		AuthCookie(c *fiber.Ctx) error
		DeleteSession(session *session.Session)
		IsLoggedIn(c *fiber.Ctx) bool
		Logout(c *fiber.Ctx) error
		IsAdmin(c *fiber.Ctx) bool
		AccessTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error
		RefreshTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error
	}
	IAdapterService interface {
		DeleteSession(session *session.Session)
		SendPasswordResetEmail(email string, baseURL string)
		SendConfirmationEmail(email string, baseURL string)
		User(c *fiber.Ctx) (*types.User, error)
		UserID(c *fiber.Ctx) (uint, error)
		Login(c *fiber.Ctx, user *types.User) error
		IsLoggedIn(c *fiber.Ctx) bool
		AuthCookie(c *fiber.Ctx) error
		Logout(c *fiber.Ctx) error
		IsAdmin(c *fiber.Ctx) bool
		AccessTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error
		RefreshTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error
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

func (adapter *accountAdapterService) User(c *fiber.Ctx) (*types.User, error) {
	return adapter.adapterType.User(c)
}

func (adapter *accountAdapterService) IsAdmin(c *fiber.Ctx) bool {
	return adapter.adapterType.IsAdmin(c)
}

func (adapter *accountAdapterService) UserID(c *fiber.Ctx) (uint, error) {
	return adapter.adapterType.UserID(c)
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

func (adapter *accountAdapterService) AuthCookie(c *fiber.Ctx) error {
	return adapter.adapterType.AuthCookie(c)
}

func (adapter *accountAdapterService) DeleteSession(session *session.Session) {
	adapter.adapterType.DeleteSession(session)
}

func (adapter *accountAdapterService) RefreshTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error {
	return adapter.adapterType.RefreshTokenCreate(c, user, session)
}

func (adapter *accountAdapterService) AccessTokenCreate(c *fiber.Ctx, user *types.User, session *session.Session) error {
	return adapter.adapterType.AccessTokenCreate(c, user, session)
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

func (adapter AccountAdapterService) User(context *fiber.Ctx) (*types.User, error) {
	userID, errID := adapter.UserID(context)
	if userID == 0 || errID != nil {
		return nil, errors.New("User Not Logged In")
	}
	user := &types.User{}
	_, errUser := adapter.ViewByID(userID, user)

	if errUser != nil {
		return nil, errUser
	}
	return user, nil
}

func (adapter AccountAdapterService) IsAdmin(context *fiber.Ctx) bool {
	user, err := adapter.User(context)
	if err != nil {
		return false
	}
	return user.IsAdmin
}

func (adapter AccountAdapterService) UserID(context *fiber.Ctx) (uint, error) {
	token := context.Cookies("access-token")
	accessClaims, err := adapter.IJWT.ParseAccessToken(token)
	if err != nil {
		return 0, errors.New("Unauthorized user")
	}
	session, err := app.Http.Session.Get(context)
	if err != nil {
		return 0, err
	}

	userID, ok := session.Get("user_id").(uint)
	if !ok {
		return 0, err
	}
	return userID, nil
}

func (adapter AccountAdapterService) IsLoggedIn(context *fiber.Ctx) bool {
	session, err := app.Http.Session.Get(context)
	userID := session.Get("user_id")

	if userID == nil || err != nil {
		adapter.DeleteSession(session)
		context.ClearCookie()
		return false
	}

	token := context.Cookies("access-token")
	if token == "" {
		tokenHash := session.Get("user_access-token")
		context.Cookie(&fiber.Cookie{
			Name:     "access-token",
			Value:    fmt.Sprintf("%s", tokenHash),
			Secure:   false,
			HTTPOnly: true,
		})
	}

	return true
}

func (adapter AccountAdapterService) AccessTokenCreate(context *fiber.Ctx, user *types.User, store *session.Session) error {
	accessClaims := config.NewUserClaims(
		*user,
		int64(float64(adapter.IJWT.GetAPIConfig().Expire/4)),
	)
	accesstoken, err := adapter.NewAccessToken(accessClaims)
	if err != nil {
		return err
	}

	expire, err := accessClaims.GetExpirationTime()

	if err == nil {
		store.Set("user_access-token", accesstoken)
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

func (adapter AccountAdapterService) RefreshTokenCreate(context *fiber.Ctx, user *types.User, session *session.Session) error {
	refreshClaims := config.NewUserClaims(
		*user,
		adapter.IJWT.GetAPIConfig().Expire*24*7,
	)
	refreshtoken, err := adapter.NewRefreshToken(refreshClaims)
	if err != nil {
		return err
	}
	expire, err := refreshClaims.GetExpirationTime()

	if err == nil {
		session.Set("user_refresh-token", refreshtoken)
		context.Cookie(&fiber.Cookie{
			Name:        "refresh-token",
			Value:       refreshtoken,
			Secure:      false,
			HTTPOnly:    true,
			SessionOnly: false,
			Expires:     expire.Time,
		})
	} else {
		return err
	}
	return nil
}

func (adapter AccountAdapterService) Login(context *fiber.Ctx, user *types.User) error {
	session, errS := app.Http.Session.Get(context) // get/create new session
	session.SetExpiry(time.Duration(adapter.IJWT.GetAPIConfig().Expire*24*7) * time.Second)
	if errS != nil {
		return errS
	}

	session.Set("user_id", user.ID) // save to storage
	context.Locals("user_id", user.ID)

	if err := adapter.AccessTokenCreate(context, user, session); err != nil {
		return err
	}

	if err := adapter.RefreshTokenCreate(context, user, session); err != nil {
		return err
	}

	if err := session.Save(); err != nil {
		return err
	}
	return nil
}

func (adapter AccountAdapterService) Logout(context *fiber.Ctx) error {
	session, errS := app.Http.Session.Get(context) // get/create new session
	if errS != nil {
		return errS
	}

	adapter.DeleteSession(session)

	context.ClearCookie()
	context.Set("X-DNS-Prefetch-Control", "off")
	context.Set("Pragma", "no-cache")
	context.Set("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
	context.Set("Cache-Control", "no-cache, must-revalidate, no-store, max-age=0, private")
	return nil
}

func (adapter AccountAdapterService) AuthCookie(context *fiber.Ctx) error {
	adapter.IsLoggedIn(context)
	return context.Next()
}

func (adapter AccountAdapterService) DeleteSession(session *session.Session) {
	if errDel := session.Destroy(); errDel != nil {
		panic(errDel.Error())
	}
	//		if err := session.Save(); err != nil {
	//		panic(err)
	//	}
}
