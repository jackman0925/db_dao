package db_dao

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// QueryerContext is an interface for sqlx.QueryerContext.
type QueryerContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

// ExecerContext is an interface for sqlx.ExecerContext.
type ExecerContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Executor is an interface that combines QueryerContext and ExecerContext.
type Executor interface {
	QueryerContext
	ExecerContext
}

// IDAO 定义了所有DAO方法，这是业务逻辑应该依赖的接口。
type IDAO[T any] interface {
	Get(context.Context, GetEndPoint[T]) error
	Select(context.Context, SelectEndPoint[T]) error
	Paginate(context.Context, PageEndPoint[T]) (int64, error)
	Insert(context.Context, InsertEndpoint[T]) (int64, error)
	BatchInsert(context.Context, BatchInsertEndpoint[T]) (int64, error)
	Update(context.Context, UpdateEndPoint[T]) (int64, error)
	Delete(context.Context, DeleteEndPoint[T]) (int64, error)
	BeginTx(ctx context.Context, opts ...*sql.TxOptions) (IDAO[T], error)
	Commit() error
	Rollback() error
	GetExecutor() Executor
}
