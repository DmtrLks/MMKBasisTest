package task

import (
	"testing"

	"mmktestbasisByDGanichev/internal/team"
)

func TestCanUpdateTask(t *testing.T) {
	const (
		ownerID    int64 = 1
		adminID    int64 = 2
		creatorID  int64 = 3
		assigneeID int64 = 4
		memberID   int64 = 5
	)

	status := TaskStatusDone
	title := "updated title"
	newAssigneeID := memberID

	current := Task{
		ID:         10,
		TeamID:     20,
		CreatedBy:  creatorID,
		AssigneeID: int64Pointer(assigneeID),
	}

	tests := []struct {
		name    string
		member  team.Member
		request UpdateRequest
		want    bool
	}{
		{
			name: "owner can update any field",
			member: team.Member{
				UserID: ownerID,
				Role:   team.RoleOwner,
			},
			request: UpdateRequest{Title: &title},
			want:    true,
		},
		{
			name: "admin can update any field",
			member: team.Member{
				UserID: adminID,
				Role:   team.RoleAdmin,
			},
			request: UpdateRequest{AssigneeID: &newAssigneeID},
			want:    true,
		},
		{
			name: "task creator can update any field",
			member: team.Member{
				UserID: creatorID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{Title: &title},
			want:    true,
		},
		{
			name: "assignee can update status",
			member: team.Member{
				UserID: assigneeID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{Status: &status},
			want:    true,
		},
		{
			name: "assignee cannot update title",
			member: team.Member{
				UserID: assigneeID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{Title: &title},
			want:    false,
		},
		{
			name: "assignee cannot update status and title together",
			member: team.Member{
				UserID: assigneeID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{
				Status: &status,
				Title:  &title,
			},
			want: false,
		},
		{
			name: "assignee cannot reassign task",
			member: team.Member{
				UserID: assigneeID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{AssigneeID: &newAssigneeID},
			want:    false,
		},
		{
			name: "unrelated member cannot update task",
			member: team.Member{
				UserID: memberID,
				Role:   team.RoleMember,
			},
			request: UpdateRequest{Status: &status},
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := canUpdateTask(test.member, current, test.request)
			if got != test.want {
				t.Fatalf("canUpdateTask() = %t, want %t", got, test.want)
			}
		})
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
