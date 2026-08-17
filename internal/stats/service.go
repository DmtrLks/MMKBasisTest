package stats

import (
	"context"
	"errors"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/team"
)

type reportRepository interface {
	Report(ctx context.Context, db database.DBTX, teamID int64) (Report, error)
}

type teamMemberFinder interface {
	FindMember(
		ctx context.Context,
		db database.DBTX,
		teamID int64,
		userID int64,
	) (team.Member, error)
}

type Service struct {
	db      database.DBTX
	reports reportRepository
	members teamMemberFinder
}

func NewService(
	db database.DBTX,
	reports reportRepository,
	members teamMemberFinder,
) *Service {
	return &Service{
		db:      db,
		reports: reports,
		members: members,
	}
}

func (s *Service) Report(
	ctx context.Context,
	currentUserID int64,
	teamID int64,
) (Report, error) {
	if currentUserID <= 0 {
		return Report{}, ErrInvalidUserID
	}

	if teamID <= 0 {
		return Report{}, ErrInvalidTeamID
	}

	member, err := s.members.FindMember(ctx, s.db, teamID, currentUserID)
	if errors.Is(err, team.ErrMemberNotFound) {
		return Report{}, ErrForbidden
	}

	if err != nil {
		return Report{}, err
	}

	if member.Role != team.RoleOwner && member.Role != team.RoleAdmin {
		return Report{}, ErrForbidden
	}

	return s.reports.Report(ctx, s.db, teamID)
}
