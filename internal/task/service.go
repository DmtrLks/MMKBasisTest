package task

import (
	"context"
	"errors"
	"log"
	"time"

	"mmktestbasisByDGanichev/internal/database"
	"mmktestbasisByDGanichev/internal/history"
	"mmktestbasisByDGanichev/internal/team"
)

type taskRepository interface {
	Create(ctx context.Context, db database.DBTX, task Task) (Task, error)

	List(ctx context.Context, db database.DBTX, params TaskParams) ([]Task, error)

	FindByID(ctx context.Context, db database.DBTX, taskID int64) (Task, error)

	Update(ctx context.Context, db database.DBTX, task Task, expectedVersion int64) (Task, error)
}

type historyAppender interface {
	Append(ctx context.Context, db database.DBTX, entry history.Entry) (history.Entry, error)
}

type teamMemberFinder interface {
	FindMember(
		ctx context.Context,
		db database.DBTX,
		teamID int64,
		userID int64,
	) (team.Member, error)

	IsMember(ctx context.Context, db database.DBTX, teamID int64, userID int64) (bool, error)
}

type validator interface {
	NormalizeAndValidateCreate(currentUserID int64, request *CreateRequest) error

	ValidateList(currentUserID int64, params TaskParams) error

	NormalizeAndValidateUpdate(currentUserID int64, taskID int64, request *UpdateRequest) error
}

type taskListCache interface {
	Version(ctx context.Context, teamID int64) (int64, error)

	Get(ctx context.Context, params TaskParams, version int64) ([]Response, error)

	Set(ctx context.Context, params TaskParams, version int64, response []Response) error

	Invalidate(ctx context.Context, teamID int64) error
}

type Service struct {
	db        database.Database
	tasks     taskRepository
	histories historyAppender
	members   teamMemberFinder
	validator validator
	cache     taskListCache
}

func NewService(
	db database.Database,
	tasks taskRepository,
	histories historyAppender,
	members teamMemberFinder,
	validator validator,
	cache taskListCache,
) *Service {
	return &Service{
		db:        db,
		tasks:     tasks,
		histories: histories,
		members:   members,
		validator: validator,
		cache:     cache,
	}
}

func (s *Service) Create(
	ctx context.Context,
	currentUserID int64,
	request CreateRequest,
) (Response, error) {
	if err := s.validator.NormalizeAndValidateCreate(
		currentUserID,
		&request,
	); err != nil {
		return Response{}, err
	}

	now := time.Now().UTC()

	newTask := Task{
		TeamID:      request.TeamID,
		Title:       request.Title,
		Description: request.Description,
		Status:      request.Status,
		CreatedBy:   currentUserID,
		AssigneeID:  request.AssigneeID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}

	if newTask.Status == TaskStatusDone {
		newTask.ClosedAt = &now
	}

	var created Task

	err := s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		if _, err := s.members.FindMember(ctx, tx, request.TeamID, currentUserID); err != nil {
			if errors.Is(err, team.ErrMemberNotFound) {
				return ErrNotTeamMember
			}

			return err
		}

		if request.AssigneeID != nil &&
			*request.AssigneeID != currentUserID {
			if _, err := s.members.FindMember(
				ctx,
				tx,
				request.TeamID,
				*request.AssigneeID,
			); err != nil {
				if errors.Is(err, team.ErrMemberNotFound) {
					return ErrAssigneeNotTeam
				}

				return err
			}
		}

		var err error

		created, err = s.tasks.Create(ctx, tx, newTask)
		if err != nil {
			return err
		}

		_, err = s.histories.Append(ctx, tx, history.Entry{
			TaskID:    created.ID,
			ChangedBy: currentUserID,
			Changes:   creationChanges(created),
			CreatedAt: now,
		})

		return err
	})
	if err != nil {
		return Response{}, err
	}

	s.invalidateTaskListCache(ctx, created.TeamID)

	return responseFromTask(created), nil
}

func creationChanges(task Task) history.Changes {
	return history.Changes{
		"team_id": {
			Old: nil,
			New: task.TeamID,
		},
		"title": {
			Old: nil,
			New: task.Title,
		},
		"description": {
			Old: nil,
			New: task.Description,
		},
		"status": {
			Old: nil,
			New: task.Status,
		},
		"created_by": {
			Old: nil,
			New: task.CreatedBy,
		},
		"assignee_id": {
			Old: nil,
			New: task.AssigneeID,
		},
		"closed_at": {
			Old: nil,
			New: task.ClosedAt,
		},
		"version": {
			Old: nil,
			New: task.Version,
		},
	}
}

func (s *Service) List(
	ctx context.Context,
	currentUserID int64,
	params TaskParams,
) ([]Response, error) {
	if err := s.validator.ValidateList(currentUserID, params); err != nil {
		return nil, err
	}

	isMember, err := s.members.IsMember(ctx, s.db, params.TeamID, currentUserID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrNotTeamMember
	}

	cacheVersion := int64(0)
	cacheAvailable := s.cache != nil

	if cacheAvailable {
		version, err := s.cache.Version(ctx, params.TeamID)
		if err != nil {
			log.Printf("get task list cache version: %v", err)
			cacheAvailable = false
		} else {
			cacheVersion = version

			cached, err := s.cache.Get(ctx, params, version)
			if err != nil {
				log.Printf("get task list cache: %v", err)
				cacheAvailable = false
			} else if cached != nil {
				return cached, nil
			}
		}
	}

	tasks, err := s.tasks.List(ctx, s.db, params)
	if err != nil {
		return nil, err
	}

	response := make([]Response, 0, len(tasks))

	for _, item := range tasks {
		response = append(response, responseFromTask(item))
	}

	if cacheAvailable {
		if err := s.cache.Set(
			ctx,
			params,
			cacheVersion,
			response,
		); err != nil {
			log.Printf("set task list cache: %v", err)
		}
	}

	return response, nil
}

func (s *Service) Update(
	ctx context.Context,
	currentUserID int64,
	taskID int64,
	request UpdateRequest,
) (Response, error) {
	if err := s.validator.NormalizeAndValidateUpdate(
		currentUserID,
		taskID,
		&request,
	); err != nil {
		return Response{}, err
	}

	var updated Task

	err := s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		current, err := s.tasks.FindByID(ctx, tx, taskID)
		if err != nil {
			return err
		}

		member, err := s.members.FindMember(ctx, tx, current.TeamID, currentUserID)
		if err != nil {
			if errors.Is(err, team.ErrMemberNotFound) {
				return ErrNotTeamMember
			}

			return err
		}

		if !canUpdateTask(member, current, request) {
			return ErrUpdateForbidden
		}

		if request.Version != current.Version {
			return ErrVersionConflict
		}

		if request.AssigneeID != nil &&
			*request.AssigneeID != currentUserID {
			if _, err := s.members.FindMember(
				ctx,
				tx,
				current.TeamID,
				*request.AssigneeID,
			); err != nil {
				if errors.Is(err, team.ErrMemberNotFound) {
					return ErrAssigneeNotTeam
				}

				return err
			}
		}

		now := time.Now().UTC()

		var changes history.Changes

		updated, changes = applyUpdate(current, request, now)

		if len(changes) == 0 {
			return ErrNoChanges
		}

		updated.UpdatedAt = now
		updated.Version = current.Version + 1
		changes["version"] = history.FieldChange{
			Old: current.Version,
			New: updated.Version,
		}

		updated, err = s.tasks.Update(ctx, tx, updated, current.Version)
		if err != nil {
			return err
		}

		_, err = s.histories.Append(ctx, tx, history.Entry{
			TaskID:    updated.ID,
			ChangedBy: currentUserID,
			Changes:   changes,
			CreatedAt: now,
		})

		return err
	})
	if err != nil {
		return Response{}, err
	}

	s.invalidateTaskListCache(ctx, updated.TeamID)

	return responseFromTask(updated), nil
}

func (s *Service) invalidateTaskListCache(
	ctx context.Context,
	teamID int64,
) {
	if s.cache == nil {
		return
	}

	if err := s.cache.Invalidate(ctx, teamID); err != nil {
		log.Printf("invalidate task list cache for team %d: %v", teamID, err)
	}
}

func canUpdateTask(
	member team.Member,
	current Task,
	request UpdateRequest,
) bool {
	if member.Role == team.RoleOwner || member.Role == team.RoleAdmin {
		return true
	}

	if current.CreatedBy == member.UserID {
		return true
	}

	if current.AssigneeID == nil || *current.AssigneeID != member.UserID {
		return false
	}

	return request.Status != nil &&
		request.Title == nil &&
		request.Description == nil &&
		request.AssigneeID == nil
}

func applyUpdate(
	current Task,
	request UpdateRequest,
	now time.Time,
) (Task, history.Changes) {
	updated := current
	changes := make(history.Changes)

	if request.Title != nil && *request.Title != current.Title {
		updated.Title = *request.Title
		changes["title"] = history.FieldChange{
			Old: current.Title,
			New: updated.Title,
		}
	}

	if request.Description != nil &&
		*request.Description != current.Description {
		updated.Description = *request.Description
		changes["description"] = history.FieldChange{
			Old: current.Description,
			New: updated.Description,
		}
	}

	if request.Status != nil && *request.Status != current.Status {
		updated.Status = *request.Status
		changes["status"] = history.FieldChange{
			Old: current.Status,
			New: updated.Status,
		}

		oldClosedAt := updated.ClosedAt

		if updated.Status == TaskStatusDone {
			updated.ClosedAt = &now
		} else {
			updated.ClosedAt = nil
		}

		changes["closed_at"] = history.FieldChange{
			Old: oldClosedAt,
			New: updated.ClosedAt,
		}
	}

	if request.AssigneeID != nil &&
		!equalInt64Pointers(request.AssigneeID, current.AssigneeID) {
		updated.AssigneeID = request.AssigneeID
		changes["assignee_id"] = history.FieldChange{
			Old: current.AssigneeID,
			New: updated.AssigneeID,
		}
	}

	return updated, changes
}

func equalInt64Pointers(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return *left == *right
}
