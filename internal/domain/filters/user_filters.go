// Package filters contains typed filter structs for all list/search operations.
// Handlers parse query params into these structs; usecases pass them to repos.
// This keeps filter logic in one place rather than scattered across all three layers.
package filters

// UserFilters controls the list/search operation for User records.
type UserFilters struct {
	Status   string // filter by ActorStatus (e.g., "active", "pending")
	Search   string // partial match on name or email
	Page     int    // 1-based page number
	Limit    int    // items per page (max enforced in repo)
	SortBy   string // column name to sort by (default: created_at)
	SortDir  string // "asc" | "desc"
}

func (f *UserFilters) ApplyDefaults() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortDir != "asc" && f.SortDir != "desc" {
		f.SortDir = "desc"
	}
}

func (f *UserFilters) Offset() int {
	return (f.Page - 1) * f.Limit
}
