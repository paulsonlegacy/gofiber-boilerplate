package response

import (
	"boilerplate/internal/domain/entity"
	"time"
)

// UserResponse is the public-facing representation of a User.
// Never return entity structs directly — always map through a response type.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	AvatarURL *string   `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToUserResponse maps a User entity to a UserResponse.
func ToUserResponse(u *entity.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Status:    string(u.Status),
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}

// ToUserList maps a slice of User entities to a list response.
func ToUserList(users []*entity.User, total int64) map[string]any {
	items := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, ToUserResponse(u))
	}
	return map[string]any{
		"items": items,
		"total": total,
	}
}
