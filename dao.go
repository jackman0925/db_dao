package db_dao

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// DAO is the main data access object
type DAO struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// NewDAO creates a new DAO
func NewDAO(db *sqlx.DB) *DAO {
	return &DAO{db: db}
}

// getExecutor returns the correct sqlx executor (db or tx)
func (d *DAO) getExecutor() sqlx.ExtContext {
	if d.tx != nil {
		return d.tx
	}
	return d.db
}

// BeginTx starts a transaction
func (d *DAO) BeginTx(ctx context.Context) (*DAO, error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &DAO{db: d.db, tx: tx}, nil
}

// Commit commits the transaction
func (d *DAO) Commit() error {
	if d.tx != nil {
		return d.tx.Commit()
	}
	return nil
}

// Rollback rollbacks the transaction
func (d *DAO) Rollback() error {
	if d.tx != nil {
		return d.tx.Rollback()
	}
	return nil
}

// Get executes a get query
func (d *DAO) Get(endpoint GetEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	executor := d.getExecutor()
	return executor.Get(endpoint.Model, executor.Rebind(query), args...)
}

// Select executes a select query
func (d *DAO) Select(endpoint SelectEndPoint[T]) error {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return err
	}
	executor := d.getExecutor()
	return executor.Select(endpoint.Model, executor.Rebind(query), args...)
}

// Paginate executes a paginated query
func (d *DAO) Paginate(endpoint PageEndPoint[T]) (int32, error) {
	var total int32
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	executor := d.getExecutor()
	if err := executor.Get(&total, executor.Rebind(query), args...); err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}

	query, args, err = endpoint.point2pageSql()
	if err != nil {
		return 0, err
	}
	err = executor.Select(endpoint.Model, executor.Rebind(query), args...)
	return total, err
}

// Insert executes an insert query
func (d *DAO) Insert(endpoint InsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	executor := d.getExecutor()
	result, err := executor.Exec(executor.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// BatchInsert executes a batch insert query
func (d *DAO) BatchInsert(endpoint BatchInsertEndpoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	executor := d.getExecutor()
	result, err := executor.Exec(executor.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Update executes an update query
func (d *DAO) Update(endpoint UpdateEndPoint[T]) (int64, error) {
	query, rowsArgs, conditionsArgs, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	args := append(rowsArgs, conditionsArgs...)
	executor := d.getExecutor()
	result, err := executor.Exec(executor.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete executes a delete query
func (d *DAO) Delete(endpoint DeleteEndPoint[T]) (int64, error) {
	query, args, err := endpoint.point2Sql()
	if err != nil {
		return 0, err
	}
	executor := d.getExecutor()
	result, err := executor.Exec(executor.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Exec executes a raw SQL query
func (d *DAO) Exec(query string, args ...any) (int64, error) {
	executor := d.getExecutor()
	result, err := executor.Exec(executor.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// QueryToSlice executes a raw query and maps the result to a slice
func (d *DAO) QueryToSlice[T any](query string, dest *[]T) error {
	executor := d.getExecutor()
	rows, err := executor.Queryx(executor.Rebind(query))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item T
		if err := rows.StructScan(&item); err != nil {
			return err
		}
		*dest = append(*dest, item)
	}

	return rows.Err()
}
