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

type tokenRes struct {
	Access  AccesstokenRes  `json:"access" binding:"required"`
	Refresh RefreshtokenRes `json:"refresh" binding:"required"`
}

//type UserAuth struct {
//router fiber.Router
//user   user.IUserService
//}

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

func (controller *authController) VerifyRegisteredEmail() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "User successfully verified...",
		})
	}
}

func (controller *authController) ResendConfirmEmail() fiber.Handler {
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

func (controller *authController) RequestPasswordRest() fiber.Handler {
	return func(context *fiber.Ctx) error {
		request_password_reset := &types.RequestPasswordResetForm{}
		if errForm := context.BodyParser(request_password_reset); errForm != nil {
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

func (controller *authController) PasswordReset() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controlller *authController) PasswordResetComfirm() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controller *authController) PasswordResetComplete() fiber.Handler {
	return func(context *fiber.Ctx) error {
		return context.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{},
		})
	}
}

func (controller *authController) Login() fiber.Handler {
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
		accessClaims, errAC := controller.IJWT.ParseAccessToken(access)
		if errAC != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}
		accessExpire, err := accessClaims.GetExpirationTime()
		if err != nil {
			return context.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Parsing token error"})
		}

		refresh := context.Cookies("refresh-token")
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
	controller.Auth.router.Post(
		"register",
		loginMiddleWare.RedirectToHomePageOnLogin,
		registerMiddleWare.ValidateRegister,
		controller.register(),
	)

	controller.Auth.router.Get(
		"verify-email",
		registerMiddleWare.ValidateConfirmToken,
		controller.VerifyRegisteredEmail(),
	)

	controller.Auth.router.Get(
		"resend/confirm",
		controller.ResendConfirmEmail(),
	)
	controller.Auth.router.Post(
		"request-password-reset",
		controller.RequestPasswordRest())

	controller.Auth.router.Get("reset-password",
		passwordMiddleWare.ValidatePasswordReset,
		controller.PasswordReset(),
	)
	controller.Auth.router.Post("password-reset-comfirm",
		passwordMiddleWare.ValidatePasswordResetToken,
		controller.PasswordResetComfirm(),
	)
	controller.Auth.router.Patch("reset-password-complete",
		passwordMiddleWare.ValidatePasswordResetData,
		controller.PasswordResetComplete(),
	)

	return controller
}
