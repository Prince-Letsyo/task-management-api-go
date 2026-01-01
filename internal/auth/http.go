package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
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

type httpConfiguration func(acs *httpController) error

type httpController struct {
	auth IAuthController
}

func newHTTPController(cfgs ...httpConfiguration) (*httpController, error) {
	acs := &httpController{}

	for _, cfg := range cfgs {
		if err := cfg(acs); err != nil {
			return nil, err
		}
	}

	return acs, nil
}

func withAuthRepository(auth Auth, appConfig *config.AppCfg) httpConfiguration {
	ahc := newAuthController(auth, appConfig)
	return authRepository(ahc)
}

func authRepository(controller IAuthController) httpConfiguration {
	return func(httpController *httpController) error {
		httpController.auth = controller
		return nil
	}
}
