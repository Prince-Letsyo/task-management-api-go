package profile

import "github.com/Prince-Letsyo/task-management-api-go/config"

type httpProfileController struct {
	userProfile IProfileController
}

type httpProfileConfiguration func(pcs *httpProfileController) error

func newHTTPProfileController(cfgs ...httpProfileConfiguration) (*httpProfileController, error) {
	pcs := &httpProfileController{}

	for _, cfg := range cfgs {
		if err := cfg(pcs); err != nil {
			return nil, err
		}
	}

	return pcs, nil
}

func withProfileController(userProfile Profile, appCfg *config.AppCfg) httpProfileConfiguration {
	return func(pcs *httpProfileController) error {
		pcs.userProfile = newProfileController(userProfile, appCfg)
		return nil
	}
}
