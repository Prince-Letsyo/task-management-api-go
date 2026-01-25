package profile

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

type ProfileService struct {
	IProfileRepository
	*config.AppCfg
}

func (ps *ProfileService) View(userID uint, profile *types.Profile) (*types.Profile, error) {
	p, err := ps.retrieve(userID, profile)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (ps *ProfileService) Modify(id uint, profile *types.Profile) (*types.Profile, error) {
	p, err := ps.update(id, profile)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type ProfileServiceConfiguration func(ps *ProfileService) error

func withDatabaseProfileRepository(profileRepository IProfileRepository) ProfileServiceConfiguration {
	return func(ps *ProfileService) error {
		ps.IProfileRepository = profileRepository
		return nil
	}
}

func newProfileService(appCfg *config.AppCfg, cfgs ...ProfileServiceConfiguration) (*ProfileService, error) {
	ps := &ProfileService{
		AppCfg: appCfg,
	}

	for _, cfg := range cfgs {
		if err := cfg(ps); err != nil {
			return nil, err
		}
	}
	return ps, nil
}
