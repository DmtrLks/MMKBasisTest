package team

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

type Member struct {
	TeamID int64
	UserID int64
	Role   Role
}

type TeamWithRole struct {
	Team Team
	Role Role
}
