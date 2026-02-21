package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

// CollectOne scans a single row into T using struct field name matching.
// Returns nil, nil if no rows are found.
func CollectOne[T any](rows pgx.Rows, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	v, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &v, nil
}

// CollectAll scans all rows into a slice of T using struct field name matching.
// Returns an empty slice if no rows are found.
func CollectAll[T any](rows pgx.Rows, err error) ([]T, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	v, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, err
	}

	return v, nil
}

// CollectOneScalar scans a single scalar value from one row.
// Returns the zero value of T if no rows are found.
func CollectOneScalar[T any](rows pgx.Rows, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}
	defer rows.Close()

	v, err := pgx.CollectOneRow(rows, pgx.RowTo[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, nil
		}
		return zero, err
	}

	return v, nil
}
