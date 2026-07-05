package pagination

import "github.com/google/uuid"

// PageMeta describes bidirectional cursor pagination for a page of items.
type PageMeta struct {
	HasPreviousPage bool
	HasNextPage     bool
	StartCursor     *uuid.UUID
	EndCursor       *uuid.UUID
}

// MetaInput configures how page metadata is built from a result set.
type MetaInput struct {
	HasPreviousPage bool
	HasNextPage     bool
	StartCursor     *uuid.UUID
	EndCursor       *uuid.UUID
}

// BuildMeta constructs page metadata from navigation flags and boundary cursors.
func BuildMeta(in MetaInput) PageMeta {
	return PageMeta{
		HasPreviousPage: in.HasPreviousPage,
		HasNextPage:     in.HasNextPage,
		StartCursor:     in.StartCursor,
		EndCursor:       in.EndCursor,
	}
}

// BoundaryCursors returns cursors for the first and last item IDs in items.
func BoundaryCursors(items []uuid.UUID) (start, end *uuid.UUID) {
	if len(items) == 0 {
		return nil, nil
	}
	start = &items[0]
	end = &items[len(items)-1]
	return start, end
}

// TrimOverfetch returns up to count items and whether more exist in the fetch direction.
func TrimOverfetch[T any](items []T, count int) (page []T, hasMore bool) {
	if len(items) > count {
		return items[:count], true
	}
	return items, false
}
