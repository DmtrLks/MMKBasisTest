package comment

import "strings"

const maxContentBytes = 65_535

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateAccess(currentUserID, taskID int64) error {
	if currentUserID <= 0 {
		return ErrInvalidUserID
	}

	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	return nil
}

func (v *Validator) NormalizeAndValidateCreate(
	currentUserID int64,
	taskID int64,
	request *CreateRequest,
) error {
	if err := v.ValidateAccess(currentUserID, taskID); err != nil {
		return err
	}

	request.Content = strings.TrimSpace(request.Content)
	if request.Content == "" {
		return ErrContentRequired
	}

	if len(request.Content) > maxContentBytes {
		return ErrContentTooLong
	}

	return nil
}
