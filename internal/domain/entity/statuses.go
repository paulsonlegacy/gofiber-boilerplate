package entity

// ActorStatus represents the lifecycle state of any actor (user, staff, admin).
type ActorStatus string

const (
	ActorStatusPending   ActorStatus = "pending"   // registered but not yet approved/verified
	ActorStatusActive    ActorStatus = "active"     // fully active
	ActorStatusSuspended ActorStatus = "suspended"  // temporarily disabled by admin
	ActorStatusDeleted   ActorStatus = "deleted"    // soft-deleted — token still in DB until TTL
)
