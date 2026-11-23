package filters

// AdminFilters controls the list/search operation for Admin records.
type AdminFilters struct {
	Status      string // "active" | "suspended"
	IsSuperAdmin *bool  // nil = all, true = superadmins only, false = regular only
	Search      string // partial match on name or email
	Page        int
	Limit       int
}

func (f *AdminFilters) ApplyDefaults() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
}

func (f *AdminFilters) Offset() int {
	return (f.Page - 1) * f.Limit
}
