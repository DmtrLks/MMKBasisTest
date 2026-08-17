package team

import (
	"errors"
	"testing"
)

func TestCanChangeMemberRole(t *testing.T) {
	const (
		ownerID  int64 = 1
		adminID  int64 = 2
		memberID int64 = 3
	)

	owner := Member{TeamID: 10, UserID: ownerID, Role: RoleOwner}
	admin := Member{TeamID: 10, UserID: adminID, Role: RoleAdmin}
	member := Member{TeamID: 10, UserID: memberID, Role: RoleMember}

	tests := []struct {
		name          string
		currentMember Member
		target        Member
		want          error
	}{
		{
			name:          "owner promotes member",
			currentMember: owner,
			target:        member,
			want:          nil,
		},
		{
			name:          "owner demotes admin",
			currentMember: owner,
			target:        admin,
			want:          nil,
		},
		{
			name:          "owner cannot change own role",
			currentMember: owner,
			target:        owner,
			want:          ErrOwnerRoleImmutable,
		},
		{
			name:          "admin cannot change roles",
			currentMember: admin,
			target:        member,
			want:          ErrForbidden,
		},
		{
			name:          "admin cannot change owner role",
			currentMember: admin,
			target:        owner,
			want:          ErrForbidden,
		},
		{
			name:          "member cannot change roles",
			currentMember: member,
			target:        admin,
			want:          ErrForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := canChangeMemberRole(test.currentMember, test.target)
			if !errors.Is(got, test.want) {
				t.Fatalf("canChangeMemberRole() = %v, want %v", got, test.want)
			}
		})
	}
}
