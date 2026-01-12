package types

type Profile struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id" validate:"required"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
	User      *User   `json:"user"`
}

type IProfileService interface {
	View(userID uint, profile *Profile) (*Profile, error)
	Modify(id uint, profile *Profile) (*Profile, error)
}

type ProfileForm struct {
	Bio       *string `form:"bio"`
	AvatarURL *string `form:"avatar_url"`
}
