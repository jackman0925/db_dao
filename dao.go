package db_dao

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// DAO is the main data access object, generic over a model type T
type DAO[T any] struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewDAO creates a new DAO for a specific model type
func NewDAO[T any](db *sqlx.DB) *DAO[T] {
	return &DAO[T]{db: db}
}

// BeginTx starts a transaction
func (d *DAO[T]) BeginTx(ctx context.Context) (*DAO[T], error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &DAO[T]{db: d.db, tx: tx}, nil
}

// Commit commits the transaction
func (d *DAO[T]) Commit() error {
	if d.tx != nil {
		return d.tx.Commit()
	}
	return nil
}

// Rollback rollbacks the transaction
func (d *DAO[T]) Rollback() error {
	if d.tx != nil {
		return d.tx.Rollback()
	}
	return nil
}

// rebind applies the correct bindvar type for the driver
func (d *DAO[T]) rebind(query string) string {
	return d.db.Rebind(query)
}

// Get executes a get query
func (d *DAO[T]) Get(ctx context.Context, endpoint GetEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	if d.tx != nil {
		return d.tx.GetContext(ctx, endpoint.Model, d.rebind(query), args...)
	}
	return d.db.GetContext(ctx, endpoint.Model, d.rebind(query), args...)
}

// Select executes a select query
func (d *DAO[T]) Select(ctx context.Context, endpoint SelectEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	if d.tx != nil {
		return d.tx.SelectContext(ctx, endpoint.Model, d.rebind(query), args...)
	}
	return d.db.SelectContext(ctx, endpoint.Model, d.rebind(query), args...)
}

// Paginate executes a paginated query
func (d *DAO[T]) Paginate(ctx context.Context, endpoint PageEndPoint[T]) (int32, error) {
	var total int32
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}

	if d.tx != nil {
		if err := d.tx.GetContext(ctx, &total, d.rebind(query), args...); err != nil {
			return 0, err
		}
	} else {
		if err := d.db.GetContext(ctx, &total, d.rebind(query), args...); err != nil {
			return 0, err
		}
	}

	if total == 0 {
		return 0, nil
	}

	query, args, err = endpoint.point2pageSql()
	if err != nil {
		return 0, err
	}

	if d.tx != nil {
		err = d.tx.SelectContext(ctx, endpoint.Model, d.rebind(query), args...)
	} else {
		err = d.db.SelectContext(ctx, endpoint.Model, d.rebind(query), args...)
	}
	return total, err
}

// execContext executes a query that returns rows affected
func (d *DAO[T]) execContext(ctx context.Context, query string, args ...any) (int64, error) {
	var result sql.Result
	var err error

	if d.tx != nil {
		result, err = d.tx.ExecContext(ctx, d.rebind(query), args...)
	} else {
		result, err = d.db.ExecContext(ctx, d.rebind(query), args...)
	}

	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Insert executes an insert query
func (d *DAO[T]) Insert(ctx context.Context, endpoint InsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}

// BatchInsert executes a batch insert query
func (d *DAO[T]) BatchInsert(ctx context.Context, endpoint BatchInsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}

// Update executes an update query
func (d *DAO[T]) Update(ctx context.Context, endpoint UpdateEndPoint[T]) (int64, error) {
	query, rowsArgs, conditionsArgs, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	args := append(rowsArgs, conditionsArgs...)
	return d.execContext(ctx, query, args...)
}

// Delete executes a delete query
func (d *DAO[T]) Delete(ctx context.Context, endpoint DeleteEndPoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}

// Exec executes a raw SQL query
func (d *DAO[T]) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	return d.execContext(ctx, query, args...)
}

// QueryToSlice executes a raw query and maps the result to a slice.
func (d *DAO[T]) QueryToSlice(ctx context.Context, query string, dest *[]T, args ...any) error {
	if d.tx != nil {
		return d.tx.SelectContext(ctx, dest, d.rebind(query), args...)
	}
	return d.db.SelectContext(ctx, dest, d.rebind(query), args...)
}
