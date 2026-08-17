package team

import "time"

type CreateRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	UserID int64 `json:"user_id"`
	Role   Role  `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role Role `json:"role"`
}

type Response struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	Role      Role      `json:"role"`
}

type MemberResponse struct {
	TeamID int64 `json:"team_id"`
	UserID int64 `json:"user_id"`
	Role   Role  `json:"role"`
}

func responseFromTeam(t Team, role Role) Response {
	return Response{
		ID:        t.ID,
		Name:      t.Name,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
		Role:      role,
	}
}

func memberResponseFromMember(member Member) MemberResponse {
	return MemberResponse{
		TeamID: member.TeamID,
		UserID: member.UserID,
		Role:   member.Role,
	}
}
