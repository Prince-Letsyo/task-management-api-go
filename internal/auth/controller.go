package auth

import (
	"fmt"
	"reflect"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/middleware"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/service"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type IAuthController interface {
	login() fiber.Handler
	logout() fiber.Handler
	register() fiber.Handler
	verifyRegisteredEmail() fiber.Handler
	resendConfirmEmail() fiber.Handler
	passwordResetComfirm() fiber.Handler
	passwordResetComplete() fiber.Handler
	passwordReset() fiber.Handler
	requestPasswordReset() fiber.Handler
	createNewAccessToken() fiber.Handler
	verifyToken() fiber.Handler
}

type Auth struct {
	router          fiber.Router
	userService     types.IUserService
	registerService types.IRegisterService
	loginService    types.ILoginService
}

type AccesstokenRes struct {
	AccessToken    string `json:"access_token" binding:"required"`
	AccessTokenExp string `json:"access_token_expire"`
}

type RefreshtokenRes struct {
	RefreshToken    string `json:"refresh_token" binding:"required"`
	RefreshTokenExp string `json:"refresh_token_expire" `
}

type Token struct {
	Token string `json:"token" binding:"required"`
}

type tokenRes struct {
	Access  AccesstokenRes  `json:"access" binding:"required"`
	Refresh RefreshtokenRes `json:"refresh" binding:"required"`
}

type authController struct {
	Auth
	*config.AppCfg
}

func (controller *authController) register() fiber.Handler {
	return func(context *fiber.Ctx) error {
		register := context.Locals("register").(*types.User)

		go service.NewAccountAdapterService(
			controller.AppCfg, service.AccountAdapterService{
				IUserService: controller.userService,
			}).SendConfirmationEmail(
			register.Email,
			service.GenerateConfirmPath(context, controller.Server.Redirect),
		)

		return context.JSON(fiber.Map{
			"success": true,
			"message": "Registered successfully! Please confirm your email",
		})
	}
}

func (controller *authController) verifyRegisteredEmail() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "User successfully verified...",
		})
	}
}

func (controller *authController) resendConfirmEmail() fiber.Handler {
	return func(context *fiber.Ctx) error {
		accountAdapter := service.NewAccountAdapterService(
			controller.AppCfg,
			service.AccountAdapterService{
				IUserService: controller.userService,
			})
		user := context.Locals("user").(*types.User)

		go accountAdapter.SendConfirmationEmail(
			user.Email,
			service.GenerateConfirmPath(context, app.HTTP.Server.Redirect),
		)
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Please confirm your email",
		})
	}
}

func (controller *authController) requestPasswordReset() fiber.Handler {
	return func(context *fiber.Ctx) error {
		requestPasswordReset := &types.RequestPasswordResetForm{}
		if err := context.BodyParser(requestPasswordReset); err != nil {
			return context.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "failed parsinng request  body",
			})
		}
		if errs := pkg.ValidateStruct(requestPasswordReset); len(errs) > 0 {
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": errs,
			})
		}

		if validErr := pkg.VI.Vi.Struct(requestPasswordReset); validErr != nil {
			errorsfield := make(map[string][]string)
			pass := reflect.TypeOf(types.RequestPasswordResetForm{})

			for _, err := range validErr.(validator.ValidationErrors) {
				fieldName := err.Field()
				if field, ok := pass.FieldByName(fieldName); !ok {
					return context.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{})
				} else {
					fieldJSON := field.Tag.Get("json")
					errorsfield[fieldJSON] = append(errorsfield[fieldJSON], err.Tag())
				}
			}
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": errorsfield,
			})
		}

		user := &types.User{}
		user, err := controller.userService.ViewByEmail(requestPasswordReset.Email, user)
		if err != nil {
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": "Error on request password reset request",
			})
		}
		rediect := app.HTTP.Server.Redirect

		if rediect == "" {
			rediect = fmt.Sprintf("%s/reset-password", context.BaseURL())
		} else {
			rediect = fmt.Sprintf("%s/auth/reset-password", rediect)
		}
		go service.NewAccountAdapterService(
			controller.AppCfg,
			service.AccountAdapterService{
				IUserService: controller.userService,
			}).SendPasswordResetEmail(
			user.Email,
			rediect,
		)
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "We've sent an email for password to your registered email address",
		})
	}
}

func (controller *authController) passwordReset() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controller *authController) passwordResetComfirm() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controller *authController) passwordResetComplete() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controller *authController) login() fiber.Handler {
	return func(context *fiber.Ctx) error {
		user := context.Locals("user").(*types.User)
		errS, ok := context.Locals("err").(string)
		data := fiber.Map{}

		errToken := service.
			NewAccountAdapterService(
				controller.AppCfg,
				service.AccountAdapterService{
					IUserService: controller.userService,
						}).
			Login(context, user) //nolint:wsl

		if errToken != nil && !ok {
			data["error"] = errToken.Error()
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		} else if errS != "" {
			data["error"] = errS
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}

		access := context.Cookies("access-token")
		if access == "" {
			sess, err := app.HTTP.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			access = sess.Get("access-token").(string)
		}
		accessClaims, errAC := controller.JwtSecrets.ParseAccessToken(access)
		if errAC != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		accessExpire, err := accessClaims.GetExpirationTime()
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}

		refresh := context.Cookies("refresh-token")
		if refresh == "" {
			sess, err := controller.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			refresh = sess.Get("refresh-token").(string)
		}
		refreshClaims, errRC := controller.JwtSecrets.ParseRefreshToken(refresh)
		if errRC != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		refreshExpire, err := refreshClaims.GetExpirationTime()
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}

		data["user"] = user
		data["token"] = tokenRes{
			Access: AccesstokenRes{
				AccessToken:    access,
				AccessTokenExp: accessExpire.String(),
			},
			Refresh: RefreshtokenRes{
				RefreshToken:    refresh,
				RefreshTokenExp: refreshExpire.String(),
			},
		}
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func (controller *authController) logout() fiber.Handler {
	return func(context *fiber.Ctx) error {
		data := fiber.Map{}

		err := service.NewAccountAdapterService(
			controller.AppCfg,
			service.AccountAdapterService{
				IUserService: controller.userService,
			}).Logout(context)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusInternalServerError).JSON(data)
		}

		return context.Status(fiber.StatusNoContent).JSON(data)
	}
}

func (controller *authController) createNewAccessToken() fiber.Handler {
	return func(context *fiber.Ctx) error {
		refreshTokenRes := &RefreshtokenRes{}
		data := fiber.Map{}

		if err := context.BodyParser(refreshTokenRes); err != nil {
			data["error"] = "failed parsinng request body"
			return context.Status(fiber.StatusUnprocessableEntity).JSON(data)
		}
		if errs := pkg.ValidateStruct(refreshTokenRes); len(errs) > 0 {
			data["errors"] = errs
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		refreshTokenClaim, err := controller.JwtSecrets.ParseRefreshToken(refreshTokenRes.RefreshToken)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		if err := refreshTokenClaim.Valid(); err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		}
		user := &types.User{}
		user, err = controller.userService.ViewByEmail(refreshTokenClaim.Email, user)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusInternalServerError).JSON(data)
		}
		accountService := service.NewAccountAdapterService(
			controller.AppCfg,
			service.AccountAdapterService{
				IUserService: controller.userService,
			})

		err = accountService.AccessTokenCreate(context, user)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusInternalServerError).JSON(data)
		}
		access := context.Cookies("access-token")
		if access == "" {
			sess, err := app.HTTP.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			access = sess.Get("access-token").(string)
		}
		accessTokenClaim, err := controller.JwtSecrets.ParseAccessToken(access)
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		data["token"] = AccesstokenRes{
			AccessToken:    access,
			AccessTokenExp: accessTokenClaim.ExpiresAt.String(),
		}
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func (controller *authController) verifyToken() fiber.Handler {
	return func(context *fiber.Ctx) error {
		token := &Token{}
		data := fiber.Map{}

		if err := context.BodyParser(token); err != nil {
			data["error"] = "failed parsinng request body"
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		if errs := pkg.ValidateStruct(token); len(errs) > 0 {
			data["errors"] = errs
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		accessToken, err := controller.JwtSecrets.ParseAccessToken(token.Token)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}

		data["token"] = AccesstokenRes{
			AccessToken:    token.Token,
			AccessTokenExp: accessToken.ExpiresAt.String(),
		}

		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func newAuthController(
	authC Auth,
	appCfg *config.AppCfg,
) IAuthController {
	controller := &authController{
		Auth:   authC,
		AppCfg: appCfg,
	}
	registerMiddleWare := middleware.NewRegisterMiddleWare(
		middleware.RegisterMiddleWare{
			IUserService:     controller.userService,
			IRegisterService: controller.registerService,
		}, appCfg)

	loginMiddleWare := middleware.NewLoginMiddleWare(
		middleware.LoginMiddleWare{
			IUserService:  controller.userService,
			ILoginService: controller.loginService,
		}, appCfg)
	passwordMiddleWare := middleware.NewPasswordResetMiddleWare(
		middleware.PasswordResetMiddleWare{
			IUserService:     controller.userService,
			IRegisterService: controller.registerService,
		}, appCfg)
	controller.router.Post(
		"register",
		registerMiddleWare.ValidateRegister,
		controller.register(),
	)

	controller.router.Get(
		"verify-email",
		registerMiddleWare.ValidateConfirmToken,
		controller.verifyRegisteredEmail(),
	)

	controller.router.Get(
		"resend/confirm",
		controller.resendConfirmEmail(),
	)
	controller.router.Post(
		"request-password-reset",
		controller.requestPasswordReset())

	controller.router.Get("reset-password",
		passwordMiddleWare.ValidatePasswordReset,
		controller.passwordReset(),
	)
	controller.router.Post("password-reset-confirm",
		passwordMiddleWare.ValidatePasswordResetToken,
		controller.passwordResetComfirm(),
	)
	controller.router.Patch("reset-password-complete",
		passwordMiddleWare.ValidatePasswordResetData,
		controller.passwordResetComplete(),
	)
	controller.router.Post("login",
		loginMiddleWare.ValidateLoginPost,
		controller.login())
	controller.router.Post("logout", controller.logout())

	tokenRoute := controller.router.Group("token/",
		loginMiddleWare.RedirectToHomePageOnLogin,
	)

	tokenRoute.Post(
		"refresh/",
		controller.createNewAccessToken(),
	)
	tokenRoute.Post(
		"verify/",
		controller.verifyToken(),
	)

	return controller
}
