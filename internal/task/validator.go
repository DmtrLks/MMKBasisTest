package task

import (
	"strings"
	"unicode/utf8"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) NormalizeAndValidateCreate(
	currentUserID int64,
	request *CreateRequest,
) error {
	if currentUserID <= 0 {
		return ErrInvalidUserID
	}

	if request.TeamID <= 0 {
		return ErrInvalidTeamID
	}

	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)

	if request.Title == "" {
		return ErrTitleRequired
	}

	if utf8.RuneCountInString(request.Title) > 255 {
		return ErrTitleTooLong
	}

	if request.Status == "" {
		request.Status = TaskStatusTodo
	}

	if !request.Status.IsValid() {
		return ErrInvalidStatus
	}

	if request.AssigneeID != nil && *request.AssigneeID <= 0 {
		return ErrInvalidAssignee
	}

	return nil
}

func (v *Validator) ValidateList(
	currentUserID int64,
	params TaskParams,
) error {
	if currentUserID <= 0 {
		return ErrInvalidUserID
	}

	if params.TeamID <= 0 {
		return ErrInvalidTeamID
	}

	if params.Status != nil && !params.Status.IsValid() {
		return ErrInvalidStatus
	}

	if params.AssigneeID != nil && *params.AssigneeID <= 0 {
		return ErrInvalidAssignee
	}

	if params.Limit <= 0 || params.Limit > 100 {
		return ErrInvalidLimit
	}

	if params.Offset < 0 {
		return ErrInvalidOffset
	}

	return nil
}

func (v *Validator) NormalizeAndValidateUpdate(
	currentUserID int64,
	taskID int64,
	request *UpdateRequest,
) error {
	if currentUserID <= 0 {
		return ErrInvalidUserID
	}

	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if request.Version <= 0 {
		return ErrInvalidVersion
	}

	if request.Title == nil &&
		request.Description == nil &&
		request.Status == nil &&
		request.AssigneeID == nil {
		return ErrNoUpdateFields
	}

	if request.Title != nil {
		value := strings.TrimSpace(*request.Title)
		request.Title = &value

		if value == "" {
			return ErrTitleRequired
		}

		if utf8.RuneCountInString(value) > 255 {
			return ErrTitleTooLong
		}
	}

	if request.Description != nil {
		value := strings.TrimSpace(*request.Description)
		request.Description = &value
	}

	if request.Status != nil && !request.Status.IsValid() {
		return ErrInvalidStatus
	}

	if request.AssigneeID != nil && *request.AssigneeID <= 0 {
		return ErrInvalidAssignee
	}

	return nil
}
