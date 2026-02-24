package db_dao

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

// --- Test Suite Setup ---

type DAOTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	userDAO IDAO[User]
}

type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (s *DAOTestSuite) SetupSuite() {
	// 使用内存中的 SQLite 进行测试，速度快且无需外部依赖
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		s.T().Fatalf("failed to connect to db: %v", err)
	}
	s.db = db
	s.userDAO = NewDAO[User](s.db)
}

func (s *DAOTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *DAOTestSuite) SetupTest() {
	// 每个测试开始前，都创建一个干净的表
	_, err := s.db.Exec(`DROP TABLE IF EXISTS users`)
	s.Require().NoError(err)
	_, err = s.db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	s.Require().NoError(err)

	// 插入一些种子数据
	_, err = s.db.Exec(`INSERT INTO users (id, name, age) VALUES (1, 'Alice', 30), (2, 'Bob', 40)`)
	s.Require().NoError(err)
}

func TestDAOSuite(t *testing.T) {
	suite.Run(t, new(DAOTestSuite))
}

// --- Test Cases ---

func (s *DAOTestSuite) TestGet() {
	ctx := context.Background()
	var user User
	err := s.userDAO.Get(ctx, GetEndPoint[User]{
		Model:      &user,
		Table:      "users",
		Conditions: map[string]any{"id = ": 1},
	})

	s.NoError(err)
	s.Equal(int64(1), user.ID)
	s.Equal("Alice", user.Name)
}

func (s *DAOTestSuite) TestSelect() {
	ctx := context.Background()
	var users []User
	err := s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model: &users,
		Table: "users",
	})

	s.NoError(err)
	s.Len(users, 2)
}

func (s *DAOTestSuite) TestInsert() {
	ctx := context.Background()
	affected, err := s.userDAO.Insert(ctx, InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "Charlie", "age": 50},
	})

	s.NoError(err)
	s.Equal(int64(1), affected)

	var user User
	err = s.db.Get(&user, "SELECT * FROM users WHERE name = 'Charlie'")
	s.NoError(err)
	s.Equal(int64(3), user.ID) // sqlite auto-increments
}

func (s *DAOTestSuite) TestUpdate() {
	ctx := context.Background()
	affected, err := s.userDAO.Update(ctx, UpdateEndPoint[User]{
		Table:      "users",
		Rows:       map[string]any{"age": 31},
		Conditions: map[string]any{"id = ": 1},
	})

	s.NoError(err)
	s.Equal(int64(1), affected)

	var user User
	s.db.Get(&user, "SELECT * FROM users WHERE id = 1")
	s.Equal(31, user.Age)
}

func (s *DAOTestSuite) TestDelete() {
	ctx := context.Background()
	affected, err := s.userDAO.Delete(ctx, DeleteEndPoint[User]{
		Table:      "users",
		Conditions: map[string]any{"id = ": 1},
	})

	s.NoError(err)
	s.Equal(int64(1), affected)

	var count int
	s.db.Get(&count, "SELECT count(*) FROM users")
	s.Equal(1, count)
}

func (s *DAOTestSuite) TestPaginate() {
	ctx := context.Background()
	var users []User
	total, err := s.userDAO.Paginate(ctx, PageEndPoint[User]{
		Model:    &users,
		Table:    "users",
		PageNo:   1,
		PageSize: 1,
	})

	s.NoError(err)
	s.Equal(int64(2), total)
	s.Len(users, 1)
}

func (s *DAOTestSuite) TestTransaction_Commit() {
	ctx := context.Background()

	// 1. 开始事务
	txDAO, err := s.userDAO.BeginTx(ctx)
	s.Require().NoError(err)
	s.NotNil(txDAO)

	// 2. 在事务中执行操作
	_, err = txDAO.Insert(ctx, InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "TxUser", "age": 99},
	})
	s.Require().NoError(err)

	// 3. 提交事务
	err = txDAO.Commit()
	s.NoError(err)

	// 4. 验证数据是否已写入
	var user User
	err = s.userDAO.Get(ctx, GetEndPoint[User]{
		Model:      &user,
		Table:      "users",
		Conditions: map[string]any{"name = ": "TxUser"},
	})
	s.NoError(err)
	s.Equal(99, user.Age)
}

func (s *DAOTestSuite) TestTransaction_Rollback() {
	ctx := context.Background()

	// 1. 开始事务
	txDAO, err := s.userDAO.BeginTx(ctx)
	s.Require().NoError(err)

	// 2. 在事务中执行操作
	_, err = txDAO.Insert(ctx, InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "TxUserRollback", "age": 100},
	})
	s.Require().NoError(err)

	// 3. 回滚事务
	err = txDAO.Rollback()
	s.NoError(err)

	// 4. 验证数据是否未写入
	var user User
	err = s.userDAO.Get(ctx, GetEndPoint[User]{
		Model:      &user,
		Table:      "users",
		Conditions: map[string]any{"name = ": "TxUserRollback"},
	})
	s.Error(err) // 应该会出错，因为找不到数据
}

func (s *DAOTestSuite) TestErrorHandling() {
	// 在非事务 DAO 上调用 Commit/Rollback 应该返回错误
	err := s.userDAO.Commit()
	s.ErrorIs(err, sql.ErrTxDone)

	err = s.userDAO.Rollback()
	s.ErrorIs(err, sql.ErrTxDone)

	// 在事务 DAO 上调用 BeginTx 应该返回错误
	txDAO, _ := s.userDAO.BeginTx(context.Background())
	_, err = txDAO.BeginTx(context.Background())
	s.ErrorIs(err, sql.ErrTxDone)
}

func (s *DAOTestSuite) TestOrConditions() {
	ctx := context.Background()
	var users []User

	// Test 1: Simple OR
	// SELECT * FROM users WHERE (age = 30 OR age = 40)
	err := s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model: &users,
		Table: "users",
		Conditions: map[string]any{
			"or_group": Or{
				{"age = ": 30},
				{"age = ": 40},
			},
		},
	})
	s.NoError(err)
	s.Len(users, 2) // Alice(30) and Bob(40)

	// Test 2: Mixed AND/OR
	// SELECT * FROM users WHERE name = 'Alice' AND (age = 30 OR age = 40)
	users = nil
	err = s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model: &users,
		Table: "users",
		Conditions: map[string]any{
			"name = ": "Alice",
			"or_group": Or{
				{"age = ": 30},
				{"age = ": 40},
			},
		},
	})
	s.NoError(err)
	s.Len(users, 1)
	s.Equal("Alice", users[0].Name)
}

func (s *DAOTestSuite) TestBatchInsert() {
	ctx := context.Background()
	affected, err := s.userDAO.BatchInsert(ctx, BatchInsertEndpoint[User]{
		Table: "users",
		Rows: []map[string]any{
			{"name": "Charlie", "age": 50},
			{"name": "Diana", "age": 60},
		},
	})

	s.NoError(err)
	s.Equal(int64(2), affected)

	// Verify both records were inserted
	var users []User
	err = s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model: &users,
		Table: "users",
	})
	s.NoError(err)
	s.Len(users, 4) // 2 seed + 2 new
}

func (s *DAOTestSuite) TestSelectWithInClause() {
	ctx := context.Background()
	var users []User
	err := s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model: &users,
		Table: "users",
		Conditions: map[string]any{
			"id": []int64{1, 2},
		},
	})

	s.NoError(err)
	s.Len(users, 2)
}

func (s *DAOTestSuite) TestSelectWithFields() {
	ctx := context.Background()
	var users []User
	err := s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model:  &users,
		Table:  "users",
		Fields: []string{"id", "name", "age"},
	})

	s.NoError(err)
	s.Len(users, 2)
}

func (s *DAOTestSuite) TestSelectWithAppends() {
	ctx := context.Background()
	var users []User
	err := s.userDAO.Select(ctx, SelectEndPoint[User]{
		Model:   &users,
		Table:   "users",
		Appends: []string{"ORDER BY age DESC"},
	})

	s.NoError(err)
	s.Require().Len(users, 2)
	s.Equal("Bob", users[0].Name) // Bob(40) should come first
}

func (s *DAOTestSuite) TestPaginateWithSort() {
	ctx := context.Background()
	var users []User
	total, err := s.userDAO.Paginate(ctx, PageEndPoint[User]{
		Model:     &users,
		Table:     "users",
		PageNo:    1,
		PageSize:  1,
		SortField: "age",
		SortOrder: "DESC",
	})

	s.NoError(err)
	s.Equal(int64(2), total)
	s.Require().Len(users, 1)
	s.Equal("Bob", users[0].Name) // Bob(40) should be first when sorted DESC
}

func (s *DAOTestSuite) TestPaginateWithInvalidPageNo() {
	ctx := context.Background()
	var users []User
	_, err := s.userDAO.Paginate(ctx, PageEndPoint[User]{
		Model:    &users,
		Table:    "users",
		PageNo:   0,
		PageSize: 10,
	})

	s.Error(err)
	s.Contains(err.Error(), "pageNo")
}

func (s *DAOTestSuite) TestPaginateWithInvalidPageSize() {
	ctx := context.Background()
	var users []User
	_, err := s.userDAO.Paginate(ctx, PageEndPoint[User]{
		Model:    &users,
		Table:    "users",
		PageNo:   1,
		PageSize: 0,
	})

	s.Error(err)
	s.Contains(err.Error(), "pageSize")
}

func (s *DAOTestSuite) TestInsertEmptyTable() {
	ctx := context.Background()
	_, err := s.userDAO.Insert(ctx, InsertEndpoint[User]{
		Table: "",
		Rows:  map[string]any{"name": "Test"},
	})
	s.Error(err)
}

func (s *DAOTestSuite) TestInsertEmptyRows() {
	ctx := context.Background()
	_, err := s.userDAO.Insert(ctx, InsertEndpoint[User]{
		Table: "users",
	})
	s.Error(err)
}

func (s *DAOTestSuite) TestUpdateEmptyConditions() {
	ctx := context.Background()
	_, err := s.userDAO.Update(ctx, UpdateEndPoint[User]{
		Table: "users",
		Rows:  map[string]any{"age": 99},
	})
	s.Error(err)
}

func (s *DAOTestSuite) TestDeleteEmptyConditions() {
	ctx := context.Background()
	_, err := s.userDAO.Delete(ctx, DeleteEndPoint[User]{
		Table: "users",
	})
	s.Error(err)
}

func (s *DAOTestSuite) TestBeginTxWithOptions() {
	ctx := context.Background()
	txDAO, err := s.userDAO.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	s.Require().NoError(err)
	s.NotNil(txDAO)

	// Clean up
	err = txDAO.Rollback()
	s.NoError(err)
}

func (s *DAOTestSuite) TestGetNotFound() {
	ctx := context.Background()
	var user User
	err := s.userDAO.Get(ctx, GetEndPoint[User]{
		Model:      &user,
		Table:      "users",
		Conditions: map[string]any{"id = ": 9999},
	})
	s.Error(err) // sql.ErrNoRows
}
