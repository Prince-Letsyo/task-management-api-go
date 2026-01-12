package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
)

type httpAuthConfiguration func(acs *httpAuthController) error

type httpAuthController struct {
	auth IAuthController
}

func newHTTPAuthController(cfgs ...httpAuthConfiguration) (*httpAuthController, error) {
	acs := &httpAuthController{}

	for _, cfg := range cfgs {
		if err := cfg(acs); err != nil {
			return nil, err
		}
	}

	return acs, nil
}

func withAuthController(auth Auth, appCfg *config.AppCfg) httpAuthConfiguration {
	return func(acs *httpAuthController) error {
		acs.auth = newAuthController(auth, appCfg)
		return nil
	}
}
