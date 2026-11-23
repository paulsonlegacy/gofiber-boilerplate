package response

import (
	"boilerplate/internal/domain/entity"
	"time"
)

// AdminResponse is the public-facing representation of an Admin.
type AdminResponse struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	IsSuperAdmin bool      `json:"isSuperAdmin"`
	CreatedAt    time.Time `json:"createdAt"`
}

func ToAdminResponse(a *entity.Admin) *AdminResponse {
	return &AdminResponse{
		ID:           a.ID,
		FirstName:    a.FirstName,
		LastName:     a.LastName,
		Email:        a.Email,
		Status:       string(a.Status),
		IsSuperAdmin: a.IsSuperAdmin,
		CreatedAt:    a.CreatedAt,
	}
}

func ToAdminList(admins []*entity.Admin, total int64) map[string]any {
	items := make([]*AdminResponse, 0, len(admins))
	for _, a := range admins {
		items = append(items, ToAdminResponse(a))
	}
	return map[string]any{
		"items": items,
		"total": total,
	}
}
