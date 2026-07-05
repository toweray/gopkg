package pagination

import "github.com/google/uuid"

// CursorRequest holds optional after/before cursors from a list request.
type CursorRequest struct {
	After  *uuid.UUID
	Before *uuid.UUID
}

// ParseCursorRequest parses after/before query values.
// Returns an error when both cursors are provided or either value is invalid.
func ParseCursorRequest(afterRaw, beforeRaw string) (CursorRequest, error) {
	after, err := parseOptionalUUID(afterRaw)
	if err != nil {
		return CursorRequest{}, err
	}
	before, err := parseOptionalUUID(beforeRaw)
	if err != nil {
		return CursorRequest{}, err
	}
	if after != nil && before != nil {
		return CursorRequest{}, ErrBothCursors
	}
	return CursorRequest{After: after, Before: before}, nil
}

func parseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
