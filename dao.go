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
func (d *DAO[T]) Get(endpoint GetEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	if d.tx != nil {
		return d.tx.Get(endpoint.Model, d.rebind(query), args...)
	}
	return d.db.Get(endpoint.Model, d.rebind(query), args...)
}

// Select executes a select query
func (d *DAO[T]) Select(endpoint SelectEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	if d.tx != nil {
		return d.tx.Select(endpoint.Model, d.rebind(query), args...)
	}
	return d.db.Select(endpoint.Model, d.rebind(query), args...)
}

// Paginate executes a paginated query
func (d *DAO[T]) Paginate(endpoint PageEndPoint[T]) (int32, error) {
	var total int32
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}

	if d.tx != nil {
		if err := d.tx.Get(&total, d.rebind(query), args...); err != nil {
			return 0, err
		}
	} else {
		if err := d.db.Get(&total, d.rebind(query), args...); err != nil {
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
		err = d.tx.Select(endpoint.Model, d.rebind(query), args...)
	} else {
		err = d.db.Select(endpoint.Model, d.rebind(query), args...)
	}
	return total, err
}

// exec executes a query that returns rows affected
func (d *DAO[T]) exec(query string, args ...any) (int64, error) {
	var result sql.Result
	var err error

	if d.tx != nil {
		result, err = d.tx.Exec(d.rebind(query), args...)
	} else {
		result, err = d.db.Exec(d.rebind(query), args...)
	}

	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Insert executes an insert query
func (d *DAO[T]) Insert(endpoint InsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.exec(query, args...)
}

// BatchInsert executes a batch insert query
func (d *DAO[T]) BatchInsert(endpoint BatchInsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.exec(query, args...)
}

// Update executes an update query
func (d *DAO[T]) Update(endpoint UpdateEndPoint[T]) (int64, error) {
	query, rowsArgs, conditionsArgs, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	args := append(rowsArgs, conditionsArgs...)
	return d.exec(query, args...)
}

// Delete executes a delete query
func (d *DAO[T]) Delete(endpoint DeleteEndPoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.exec(query, args...)
}

// Exec executes a raw SQL query
func (d *DAO[T]) Exec(query string, args ...any) (int64, error) {
	return d.exec(query, args...)
}

// QueryToSlice executes a raw query and maps the result to a slice.
func (d *DAO[T]) QueryToSlice(query string, dest *[]T, args ...any) error {
	if d.tx != nil {
		return d.tx.Select(dest, d.rebind(query), args...)
	}
	return d.db.Select(dest, d.rebind(query), args...)
}