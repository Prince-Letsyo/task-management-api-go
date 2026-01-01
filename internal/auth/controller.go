package auth

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/middleware"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/service"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	router          fiber.Router
	userService     types.IUserService
	registerService types.IRegisterService
	loginService    types.ILoginService
}

type AccesstokenRes struct {
	AccessToken     string `json:"access_token" binding:"required"`
	AccessToken_Exp string `json:"access_token_expire"`
}

type RefreshtokenRes struct {
	RefreshToken     string `json:"refresh_token" binding:"required"`
	RefreshToken_Exp string `json:"refresh_token_expire" `
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
	config.IJWT
	config.JwtSecrets
	config time.Duration
}

func (controller *authController) register() fiber.Handler {
	return func(context *fiber.Ctx) error {
		register := context.Locals("register").(*types.RegisterForm)

		go service.NewAccountAdapterService(service.AccountAdapterService{
			IUserService: controller.userService,
			IJWT:         controller.IJWT,
		}).SendConfirmationEmail(
			register.Email,
			service.GenerateConfirmPath(context, app.Http.Server.Redirect),
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
			service.AccountAdapterService{
				IUserService: controller.userService,
				IJWT:         controller.IJWT,
			})
		user := context.Locals("user").(*types.User)

		go accountAdapter.SendConfirmationEmail(
			user.Email,
			service.GenerateConfirmPath(context, app.Http.Server.Redirect),
		)
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Please confirm your email",
		})
	}
}

func (controller *authController) requestPasswordReset() fiber.Handler {
	return func(context *fiber.Ctx) error {
		request_password_reset := &types.RequestPasswordResetForm{}
		if errForm := pkg.CustomBodyParser(context, request_password_reset); errForm != nil {
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   errForm.Error(),
				"message": "Error on request password reset request",
			})
		}
		if validErr := pkg.VI.Vi.Struct(request_password_reset); validErr != nil {
			errorsfield := make(map[string][]string)
			pass := reflect.TypeOf(types.RequestPasswordResetForm{})

			for _, err := range validErr.(validator.ValidationErrors) {
				fieldName := err.Field()
				if field, ok := pass.FieldByName(fieldName); !ok {
					return context.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{})
				} else {
					fieldJson := field.Tag.Get("json")
					errorsfield[fieldJson] = append(errorsfield[fieldJson], err.Tag())
				}
			}
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": errorsfield,
			})
		}

		user := &types.User{}
		_, err := controller.Auth.userService.ViewByEmail(request_password_reset.Email, user)
		if err != nil {
			return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": "Error on request password reset request",
			})
		}
		rediect := app.Http.Server.Redirect

		if rediect == "" {
			rediect = fmt.Sprintf("%s/reset-password", context.BaseURL())
		} else {
			rediect = fmt.Sprintf("%s/auth/reset-password", rediect)
		}
		go service.NewAccountAdapterService(service.AccountAdapterService{
			IUserService: controller.userService,
			IJWT:         controller.IJWT,
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
			NewAccountAdapterService(service.AccountAdapterService{
				IUserService: controller.userService,
				IJWT:         controller.IJWT,
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
			sess, err := app.Http.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			access = sess.Get("access-token").(string)
		}
		accessClaims, errAC := controller.IJWT.ParseAccessToken(access)
		if errAC != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		accessExpire, err := accessClaims.GetExpirationTime()
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}

		refresh := context.Cookies("refresh-token")
		if refresh == "" {
			sess, err := app.Http.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			refresh = sess.Get("refresh-token").(string)
		}
		refreshClaims, errRC := controller.IJWT.ParseRefreshToken(refresh)
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
				AccessToken:     access,
				AccessToken_Exp: accessExpire.String(),
			},
			Refresh: RefreshtokenRes{
				RefreshToken:     refresh,
				RefreshToken_Exp: refreshExpire.String(),
			},
		}
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func (controller *authController) logout() fiber.Handler {
	return func(context *fiber.Ctx) error {
		data := fiber.Map{}

		err := service.NewAccountAdapterService(
			service.AccountAdapterService{
				IUserService: controller.userService,
				IJWT:         controller.IJWT,
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

		if err := pkg.CustomBodyParser(context, refreshTokenRes); err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusUnprocessableEntity).JSON(data)
		}
		refreshTokenClaim, err := controller.IJWT.ParseRefreshToken(refreshTokenRes.RefreshToken)
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
		account_service := service.NewAccountAdapterService(service.AccountAdapterService{
			IUserService: controller.userService,
			IJWT:         controller.IJWT,
		})

		err = account_service.AccessTokenCreate(context, user)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusInternalServerError).JSON(data)
		}
		access := context.Cookies("access-token")
		if access == "" {
			sess, err := app.Http.Session.Get(context)
			if err != nil {
				data["error"] = "Session error"
				return context.Status(fiber.StatusInternalServerError).JSON(data)
			}
			access = sess.Get("access-token").(string)
		}
		accessTokenClaim, err := controller.IJWT.ParseAccessToken(access)
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		data["token"] = AccesstokenRes{
			AccessToken:     access,
			AccessToken_Exp: accessTokenClaim.ExpiresAt.String(),
		}
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func (controller *authController) verifyToken() fiber.Handler {
	return func(context *fiber.Ctx) error {
		token := &Token{}
		data := fiber.Map{}

		if err := pkg.CustomBodyParser(context, token); err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		accessToken, err := controller.IJWT.ParseAccessToken(token.Token)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}

		data["token"] = AccesstokenRes{
			AccessToken:     token.Token,
			AccessToken_Exp: accessToken.ExpiresAt.String(),
		}

		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func newAuthController(
	authC Auth,
	authJWT config.IJWT,
) IAuthController {
	controller := &authController{
		Auth:       authC,
		IJWT:       authJWT,
		JwtSecrets: app.Http.JwtSecrets,
		config:     app.Http.HTTP.Timeout,
	}

	registerMiddleWare := middleware.NewRegisterMiddleWare(
		middleware.RegisterMiddleWare{
			IUserService:     controller.userService,
			IRegisterService: controller.registerService,
			IJWT:             controller.IJWT,
		},
	)

	loginMiddleWare := middleware.NewLoginMiddleWare(
		middleware.LoginMiddleWare{
			IUserService:  controller.userService,
			ILoginService: controller.loginService,
			IJWT:          controller.IJWT,
		})
	passwordMiddleWare := middleware.NewPasswordResetMiddleWare(
		middleware.PasswordResetMiddleWare{
			IUserService:     controller.userService,
			IRegisterService: controller.registerService,
		})
	controller.router.Post(
		"register",
		loginMiddleWare.RedirectToHomePageOnLogin,
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
