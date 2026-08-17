package team

import (
	"context"
	"errors"
	"time"

	"mmktestbasisByDGanichev/internal/database"
)

type teamRepository interface {
	Create(ctx context.Context, db database.DBTX, team Team) (Team, error)

	AddMember(ctx context.Context, db database.DBTX, member Member) error

	ListByUserID(ctx context.Context, db database.DBTX, userID int64) ([]TeamWithRole, error)

	FindMember(ctx context.Context, db database.DBTX, teamID int64, userID int64) (Member, error)

	FindMemberForUpdate(
		ctx context.Context,
		db database.DBTX,
		teamID int64,
		userID int64,
	) (Member, error)

	UpdateMemberRole(ctx context.Context, db database.DBTX, member Member) error
}

type userExistenceChecker interface {
	ExistsByID(ctx context.Context, db database.DBTX, userID int64) (bool, error)
}

type validator interface {
	NormalizeAndValidateCreate(currentUserID int64, request *CreateRequest) error

	ValidateInvite(currentUserID int64, teamID int64, request InviteRequest) error

	ValidateUpdateMemberRole(
		currentUserID int64,
		teamID int64,
		memberUserID int64,
		request UpdateMemberRoleRequest,
	) error
}

type Service struct {
	db        database.Database
	teams     teamRepository
	users     userExistenceChecker
	validator validator
}

func NewService(
	db database.Database,
	teams teamRepository,
	users userExistenceChecker,
	validator validator,
) *Service {
	return &Service{
		db:        db,
		teams:     teams,
		users:     users,
		validator: validator,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	request CreateRequest,
) (Response, error) {
	if err := s.validator.NormalizeAndValidateCreate(
		userID,
		&request,
	); err != nil {
		return Response{}, err
	}

	newTeam := Team{
		Name:      request.Name,
		CreatedBy: userID,
		CreatedAt: time.Now().UTC(),
	}

	var created Team

	err := s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		var err error

		created, err = s.teams.Create(ctx, tx, newTeam)
		if err != nil {
			return err
		}

		return s.teams.AddMember(ctx, tx, Member{
			TeamID: created.ID,
			UserID: userID,
			Role:   RoleOwner,
		})
	})
	if err != nil {
		return Response{}, err
	}

	return responseFromTeam(created, RoleOwner), nil
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
) ([]Response, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	teams, err := s.teams.ListByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	response := make([]Response, 0, len(teams))

	for _, item := range teams {
		response = append(response, responseFromTeam(item.Team, item.Role))
	}

	return response, nil
}

func (s *Service) Invite(
	ctx context.Context,
	currentUserID int64,
	teamID int64,
	request InviteRequest,
) (MemberResponse, error) {
	if err := s.validator.ValidateInvite(
		currentUserID,
		teamID,
		request,
	); err != nil {
		return MemberResponse{}, err
	}

	member := Member{
		TeamID: teamID,
		UserID: request.UserID,
		Role:   request.Role,
	}

	err := s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		currentMember, err := s.teams.FindMember(ctx, tx, teamID, currentUserID)
		if err != nil {
			if errors.Is(err, ErrMemberNotFound) {
				return ErrForbidden
			}

			return err
		}

		if currentMember.Role != RoleOwner &&
			currentMember.Role != RoleAdmin {
			return ErrForbidden
		}

		exists, err := s.users.ExistsByID(ctx, tx, request.UserID)
		if err != nil {
			return err
		}

		if !exists {
			return ErrInvitedUserAbsent
		}

		return s.teams.AddMember(ctx, tx, member)
	})
	if err != nil {
		return MemberResponse{}, err
	}

	return memberResponseFromMember(member), nil
}

func (s *Service) UpdateMemberRole(
	ctx context.Context,
	currentUserID int64,
	teamID int64,
	memberUserID int64,
	request UpdateMemberRoleRequest,
) (MemberResponse, error) {
	if err := s.validator.ValidateUpdateMemberRole(
		currentUserID,
		teamID,
		memberUserID,
		request,
	); err != nil {
		return MemberResponse{}, err
	}

	var updated Member

	err := s.db.WithinTransaction(ctx, func(tx database.DBTX) error {
		currentMember, err := s.teams.FindMember(ctx, tx, teamID, currentUserID)
		if err != nil {
			if errors.Is(err, ErrMemberNotFound) {
				return ErrForbidden
			}

			return err
		}

		target, err := s.teams.FindMemberForUpdate(ctx, tx, teamID, memberUserID)
		if err != nil {
			return err
		}

		if err := canChangeMemberRole(currentMember, target); err != nil {
			return err
		}

		if target.Role == request.Role {
			updated = target

			return nil
		}

		updated = Member{
			TeamID: teamID,
			UserID: memberUserID,
			Role:   request.Role,
		}

		return s.teams.UpdateMemberRole(ctx, tx, updated)
	})
	if err != nil {
		return MemberResponse{}, err
	}

	return memberResponseFromMember(updated), nil
}

// canChangeMemberRole описывает, кто и кому может менять роль в команде.
// Роль самого создателя неизменна, иначе команда может остаться без owner.
func canChangeMemberRole(
	currentMember Member,
	target Member,
) error {
	if currentMember.Role != RoleOwner {
		return ErrForbidden
	}

	if target.Role == RoleOwner {
		return ErrOwnerRoleImmutable
	}

	return nil
}
