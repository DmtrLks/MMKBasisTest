package comment

import (
	"context"
	"errors"
	"time"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/task"
)

type commentRepository interface {
	Create(ctx context.Context, db database.DBTX, comment Comment) (Comment, error)

	ListByTaskID(ctx context.Context, db database.DBTX, taskID int64) ([]Comment, error)
}

type taskFinder interface {
	FindByID(ctx context.Context, db database.DBTX, taskID int64) (task.Task, error)
}

type teamMemberChecker interface {
	IsMember(ctx context.Context, db database.DBTX, teamID int64, userID int64) (bool, error)
}

type validator interface {
	ValidateAccess(currentUserID, taskID int64) error
	NormalizeAndValidateCreate(currentUserID int64, taskID int64, request *CreateRequest) error
}

type Service struct {
	db        database.DBTX
	comments  commentRepository
	tasks     taskFinder
	members   teamMemberChecker
	validator validator
}

func NewService(
	db database.DBTX,
	comments commentRepository,
	tasks taskFinder,
	members teamMemberChecker,
	validator validator,
) *Service {
	return &Service{
		db:        db,
		comments:  comments,
		tasks:     tasks,
		members:   members,
		validator: validator,
	}
}

func (s *Service) Create(
	ctx context.Context,
	currentUserID int64,
	taskID int64,
	request CreateRequest,
) (Response, error) {
	if err := s.validator.NormalizeAndValidateCreate(
		currentUserID,
		taskID,
		&request,
	); err != nil {
		return Response{}, err
	}

	if err := s.checkTaskAccess(ctx, currentUserID, taskID); err != nil {
		return Response{}, err
	}

	created, err := s.comments.Create(
		ctx,
		s.db,
		Comment{
			TaskID:    taskID,
			UserID:    currentUserID,
			Content:   request.Content,
			CreatedAt: time.Now().UTC(),
		},
	)
	if err != nil {
		return Response{}, err
	}

	return responseFromComment(created), nil
}

func (s *Service) List(
	ctx context.Context,
	currentUserID int64,
	taskID int64,
) ([]Response, error) {
	if err := s.validator.ValidateAccess(currentUserID, taskID); err != nil {
		return nil, err
	}

	if err := s.checkTaskAccess(ctx, currentUserID, taskID); err != nil {
		return nil, err
	}

	comments, err := s.comments.ListByTaskID(ctx, s.db, taskID)
	if err != nil {
		return nil, err
	}

	response := make([]Response, 0, len(comments))
	for _, item := range comments {
		response = append(response, responseFromComment(item))
	}

	return response, nil
}

func (s *Service) checkTaskAccess(
	ctx context.Context,
	currentUserID int64,
	taskID int64,
) error {
	foundTask, err := s.tasks.FindByID(ctx, s.db, taskID)
	if errors.Is(err, task.ErrTaskNotFound) {
		return ErrTaskNotFound
	}

	if err != nil {
		return err
	}

	isMember, err := s.members.IsMember(ctx, s.db, foundTask.TeamID, currentUserID)
	if err != nil {
		return err
	}

	if !isMember {
		return ErrForbidden
	}

	return nil
}
