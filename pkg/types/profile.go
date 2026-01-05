package types

type Profile struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id" validate:"required"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
	User      *User   `json:"user"`
}
