package filters

// TransactionFilters controls the list/search operation for Transaction records.
type TransactionFilters struct {
	UserID    string // restrict to a specific user (set by handler for non-admin actors)
	Direction string // "credit" | "debit"
	Type      string // transaction type constant
	Method    string // payment method constant
	Status    string // "pending" | "completed" | "failed"
	Page      int
	Limit     int
	SortBy    string
	SortDir   string
}

func (f *TransactionFilters) ApplyDefaults() {
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

func (f *TransactionFilters) Offset() int {
	return (f.Page - 1) * f.Limit
}
