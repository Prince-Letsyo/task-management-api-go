package profile

import (
	"github.com/oarkflow/log"

	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

type ProfileService struct {
	repo IProfileRepository
}

func (ps *ProfileService) View(userID uint, profile *types.Profile) (*types.Profile, error) {
	p, err := ps.repo.retrieve(userID, profile)
	if err != nil {
		log.Error().Err(err).Uint("user_id", userID).Msg("failed to view profile")
		return nil, err
	}
	return p, nil
}

func (ps *ProfileService) Modify(id uint, profile *types.Profile) (*types.Profile, error) {
	p, err := ps.repo.update(id, profile)
	if err != nil {
		log.Error().Err(err).Uint("profile_id", id).Msg("failed to modify profile")
		return nil, err
	}
	log.Info().Uint("profile_id", id).Msg("profile modified successfully")
	return p, nil
}

type ProfileServiceConfiguration func(ps *ProfileService) error

func withDatabaseProfileRepository(profileRepository IProfileRepository) ProfileServiceConfiguration {
	return func(ps *ProfileService) error {
		ps.repo = profileRepository
		return nil
	}
}

func newProfileService(cfgs ...ProfileServiceConfiguration) (*ProfileService, error) {
	ps := &ProfileService{}

	for _, cfg := range cfgs {
		if err := cfg(ps); err != nil {
			return nil, err
		}
	}
	return ps, nil
}
