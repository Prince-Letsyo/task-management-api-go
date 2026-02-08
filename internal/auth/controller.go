package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/queue"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type IAuthController interface {
	signUp() fiber.Handler
	signIn() fiber.Handler
	signInMFA() fiber.Handler
	getAccessToken() fiber.Handler
	activateAccount() fiber.Handler
	sendActivationEmail() fiber.Handler
	requestPasswordReset() fiber.Handler
	resetPassword() fiber.Handler
	logout() fiber.Handler
	logoutAll() fiber.Handler
	listSessions() fiber.Handler
	changeEmail() fiber.Handler
	adminRevokeSessions() fiber.Handler
	enable2FA() fiber.Handler
	disable2FA() fiber.Handler
}

type Auth struct {
	router fiber.Router
}

type authController struct {
	Auth
	*config.AppCfg
	authService    IAuthService
	accountService IAccountService
	queue          *queue.WorkerClient
}

func (controller *authController) signUp() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &UserCreateRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		if req.Password != req.PasswordTwo {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "passwords do not match"})
		}
		validation := pkg.DefaultValidator.ValidatePassword(req.Password, req.UserName, req.Email)
		if !validation.IsValid {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validation.Errors[0]})
		}
		_, token, _, err := controller.accountService.SignUp(req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		activationLink := controller.activationLink(c, token)
		_ = controller.queue.EnqueueEmail(queue.EmailPayload{
			To:      req.Email,
			Subject: "Activate Your Account",
			View:    "emails/confirm",
			Data: fiber.Map{
				"Title":        "Email Confirmation",
				"confirm_link": activationLink,
			},
		})
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "User created successfully. Please check your email to activate your account.",
		})
	}
}

func (controller *authController) signIn() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &AuthLoginRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		userAgent, ipAddr := requestMeta(c)
		resp, err := controller.authService.Login(req.UserName, req.Password, userAgent, ipAddr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func (controller *authController) signInMFA() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &Verify2FARequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		userAgent, ipAddr := requestMeta(c)
		resp, err := controller.authService.Login2FA(req.Token, req.TOTPToken, userAgent, ipAddr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func (controller *authController) getAccessToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &TokenRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		token, err := controller.authService.GetAccessToken(req.Token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(token)
	}
}

func (controller *authController) activateAccount() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing token"})
		}
		_, err := controller.accountService.ActivateAccount(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Account activated successfully. You can now log in."})
	}
}

func (controller *authController) sendActivationEmail() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &ActivationEmailRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		_, token, _, err := controller.accountService.SendActivationEmail(req.Email)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		activationLink := controller.activationLink(c, token)
		_ = controller.queue.EnqueueEmail(queue.EmailPayload{
			To:      req.Email,
			Subject: "Activate Your Account",
			View:    "emails/confirm",
			Data: fiber.Map{
				"Title":        "Email Confirmation",
				"confirm_link": activationLink,
			},
		})
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Activation email sent successfully. Please check your email."})
	}
}

func (controller *authController) requestPasswordReset() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &ActivationEmailRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		_, token, _, err := controller.accountService.RequestPasswordReset(req.Email)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		resetLink := controller.resetLink(c, token)
		_ = controller.queue.EnqueueEmail(queue.EmailPayload{
			To:      req.Email,
			Subject: "Password Reset Request",
			View:    "emails/password-reset",
			Data: fiber.Map{
				"Title":      "Password Reset",
				"reset_link": resetLink,
			},
		})
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "A password reset link has been sent to your email."})
	}
}

func (controller *authController) resetPassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &PasswordResetRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		if req.PasswordOne != req.PasswordTwo {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "passwords do not match"})
		}
		_, err := controller.accountService.ResetPassword(req.Token, req.Email, req.PasswordOne)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Password has been reset successfully."})
	}
}

func (controller *authController) logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := &TokenRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		if err := controller.authService.Logout(req.Token); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out"})
	}
}

func (controller *authController) logoutAll() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := AccessClaimsFromRequestWithSession(c, controller.authService)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		if err := controller.authService.LogoutAll(claims.UserID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out from all sessions"})
	}
}

func (controller *authController) listSessions() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := AccessClaimsFromRequestWithSession(c, controller.authService)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		sessions, err := controller.authService.ListSessions(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		result := make([]fiber.Map, 0, len(sessions))
		for _, session := range sessions {
			result = append(result, fiber.Map{
				"id":           session.ID,
				"user_agent":   session.UserAgent,
				"ip_address":   session.IPAddress,
				"created_at":   session.CreatedAt,
				"last_used_at": session.LastUsedAt,
				"revoked_at":   session.RevokedAt,
			})
		}
		return c.Status(fiber.StatusOK).JSON(result)
	}
}

func (controller *authController) changeEmail() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := AccessClaimsFromRequestWithSession(c, controller.authService)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		req := &ChangeEmailRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "failed parsing request body"})
		}
		if errs := pkg.ValidateStruct(req); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
		}
		if err := controller.accountService.ChangeEmail(claims.Username, req.NewEmail, req.Password); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email updated successfully"})
	}
}

func (controller *authController) adminRevokeSessions() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" || apiKey != controller.AppCfg.Auth.AdminAPIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid api key"})
		}
		userID, err := c.ParamsInt("user_id")
		if err != nil || userID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
		}
		if err := controller.authService.AdminRevokeSessions(uint(userID)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User sessions revoked"})
	}
}

func (controller *authController) enable2FA() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := AccessClaimsFromRequestWithSession(c, controller.authService)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		secret, qr, err := controller.authService.Enable2FA(claims.Username)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"secret":  secret,
			"qr_code": qr,
			"message": "2FA enabled. Scan with your app.",
		})
	}
}

func (controller *authController) disable2FA() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := AccessClaimsFromRequestWithSession(c, controller.authService)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		if err := controller.authService.Disable2FA(claims.Username); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "2FA disabled"})
	}
}

func LoadAuthRoutes(router fiber.Router, appCfg *config.AppCfg, qClient *queue.WorkerClient) {
	repo := NewDBAuthRepository(appCfg.Database.DB)
	authSvc := NewAuthService(repo, appCfg.Auth, appCfg.JwtSecrets, appCfg.Server)
	accountSvc := NewAccountService(repo, appCfg.Auth, appCfg.JwtSecrets)
	newAuthController(Auth{
		router: router.Group("/auth/"),
	}, appCfg, authSvc, accountSvc, qClient)
}

func newAuthController(authC Auth, appCfg *config.AppCfg, authSvc IAuthService, accountSvc IAccountService, qClient *queue.WorkerClient) IAuthController {
	controller := &authController{
		Auth:           authC,
		AppCfg:         appCfg,
		authService:    authSvc,
		accountService: accountSvc,
		queue:          qClient,
	}

	controller.registerRoutes()
	return controller
}

func (controller *authController) activationLink(c *fiber.Ctx, token string) string {
	base := controller.Server.Redirect
	if base == "" {
		return c.BaseURL() + "/api/auth/activate-account?token=" + token
	}
	base = strings.TrimRight(base, "/")
	return base + "/auth/activate-account?token=" + token
}

func (controller *authController) resetLink(c *fiber.Ctx, token string) string {
	base := controller.Server.Redirect
	if base == "" {
		return c.BaseURL() + "/api/auth/reset-password?token=" + token
	}
	base = strings.TrimRight(base, "/")
	return base + "/auth/reset-password?token=" + token
}

func requestMeta(c *fiber.Ctx) (*string, *string) {
	ua := c.Get("User-Agent")
	ip := c.IP()
	var uaPtr *string
	var ipPtr *string
	if ua != "" {
		uaPtr = &ua
	}
	if ip != "" {
		ipPtr = &ip
	}
	return uaPtr, ipPtr
}

func newAuthControllerWithService(authC Auth, appCfg *config.AppCfg, authSvc IAuthService, accountSvc IAccountService, qClient *queue.WorkerClient) IAuthController {
	controller := &authController{
		Auth:           authC,
		AppCfg:         appCfg,
		authService:    authSvc,
		accountService: accountSvc,
		queue:          qClient,
	}
	controller.registerRoutes()
	return controller
}

func (controller *authController) registerRoutes() {
	controller.router.Post("sign-up", controller.signUp())
	controller.router.Post("sign-in", controller.signIn())
	controller.router.Post("sign-in-mfa", controller.signInMFA())
	controller.router.Post("access", controller.getAccessToken())
	controller.router.Get("activate-account", controller.activateAccount())
	controller.router.Post("send-activation-email", controller.sendActivationEmail())
	controller.router.Post("request-password-reset", controller.requestPasswordReset())
	controller.router.Post("reset-password", controller.resetPassword())
	controller.router.Post("logout", controller.logout())
	controller.router.Post("logout-all", controller.logoutAll())
	controller.router.Get("sessions", controller.listSessions())
	controller.router.Post("change-email", controller.changeEmail())
	controller.router.Post("admin/revoke-sessions/:user_id", controller.adminRevokeSessions())
	controller.router.Post("enable-2fa", controller.enable2FA())
	controller.router.Post("disable-2fa", controller.disable2FA())
}
