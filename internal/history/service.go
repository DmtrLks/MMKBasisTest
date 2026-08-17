package history

import (
	"context"

	"mmktestbasisByDGanichev/internal/database"
)

type historyRepository interface {
	TeamIDByTaskID(ctx context.Context, db database.DBTX, taskID int64) (int64, error)

	ListByTaskID(ctx context.Context, db database.DBTX, taskID int64) ([]Entry, error)
}

type teamMemberChecker interface {
	IsMember(ctx context.Context, db database.DBTX, teamID int64, userID int64) (bool, error)
}

type Service struct {
	db        database.DBTX
	histories historyRepository
	members   teamMemberChecker
}

func NewService(
	db database.DBTX,
	histories historyRepository,
	members teamMemberChecker,
) *Service {
	return &Service{
		db:        db,
		histories: histories,
		members:   members,
	}
}

func (s *Service) List(
	ctx context.Context,
	currentUserID int64,
	taskID int64,
) ([]Response, error) {
	if currentUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	teamID, err := s.histories.TeamIDByTaskID(ctx, s.db, taskID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.members.IsMember(ctx, s.db, teamID, currentUserID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrForbidden
	}

	entries, err := s.histories.ListByTaskID(ctx, s.db, taskID)
	if err != nil {
		return nil, err
	}

	response := make([]Response, 0, len(entries))
	for _, entry := range entries {
		response = append(response, responseFromEntry(entry))
	}

	return response, nil
}
