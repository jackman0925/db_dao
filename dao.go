package db_dao

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DAO is the main data access object, generic over a model type T.
// It holds an Executor, which can be either a *sqlx.DB or a *sqlx.Tx.
type DAO[T any] struct {
	db Executor
}

// NewDAO creates a new DAO for a specific model type.
func NewDAO[T any](db Executor) *DAO[T] {
	return &DAO[T]{db: db}
}

// 确保 DAO[T] 实现了 IDAO[T] 接口
var _ IDAO[any] = (*DAO[any])(nil)

// BeginTx starts a transaction.
func (d *DAO[T]) BeginTx(ctx context.Context, opts ...*sql.TxOptions) (IDAO[T], error) {
	// Only a *sqlx.DB can begin a transaction.
	if db, ok := d.db.(*sqlx.DB); ok {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &DAO[T]{db: tx}, nil
	}
	// If it's already a transaction, return an error or handle as needed.
	return nil, sql.ErrTxDone
}

// Commit commits the transaction.
func (d *DAO[T]) Commit() error {
	if tx, ok := d.db.(*sqlx.Tx); ok {
		return tx.Commit()
	}
	return sql.ErrTxDone
}

// Rollback rollbacks the transaction.
func (d *DAO[T]) Rollback() error {
	if tx, ok := d.db.(*sqlx.Tx); ok {
		return tx.Rollback()
	}
	return sql.ErrTxDone
}

// GetExecutor returns the underlying executor.
func (d *DAO[T]) GetExecutor() Executor {
	return d.db
}

// rebind applies the correct bindvar type for the driver.
func (d *DAO[T]) rebind(query string) string {
	if db, ok := d.db.(*sqlx.DB); ok {
		return db.Rebind(query)
	}
	if tx, ok := d.db.(*sqlx.Tx); ok {
		return tx.Rebind(query)
	}
	return query
}

// Get executes a get query.
func (d *DAO[T]) Get(ctx context.Context, endpoint GetEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	return sqlx.GetContext(ctx, d.db, endpoint.Model, d.rebind(query), args...)
}

// Select executes a select query.
func (d *DAO[T]) Select(ctx context.Context, endpoint SelectEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	return sqlx.SelectContext(ctx, d.db, endpoint.Model, d.rebind(query), args...)
}

// Paginate executes a paginated query.
func (d *DAO[T]) Paginate(ctx context.Context, endpoint PageEndPoint[T]) (int64, error) {
	var total int64
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}

	if err := sqlx.GetContext(ctx, d.db, &total, d.rebind(query), args...); err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, nil
	}

	query, args, err = endpoint.point2pageSql()
	if err != nil {
		return 0, err
	}

	return total, sqlx.SelectContext(ctx, d.db, endpoint.Model, d.rebind(query), args...)
}

// execContext executes a query that returns rows affected.
func (d *DAO[T]) execContext(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := d.db.ExecContext(ctx, d.rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Insert executes an insert query.
func (d *DAO[T]) Insert(ctx context.Context, endpoint InsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}

// BatchInsert executes a batch insert query.
func (d *DAO[T]) BatchInsert(ctx context.Context, endpoint BatchInsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}

// Update executes an update query.
func (d *DAO[T]) Update(ctx context.Context, endpoint UpdateEndPoint[T]) (int64, error) {
	query, rowsArgs, conditionsArgs, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	args := append(rowsArgs, conditionsArgs...)
	return d.execContext(ctx, query, args...)
}

// Delete executes a delete query.
func (d *DAO[T]) Delete(ctx context.Context, endpoint DeleteEndPoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	return d.execContext(ctx, query, args...)
}