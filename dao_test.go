package db_dao

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// TestUser is a struct used for testing
type TestUser struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

// setupTestDAO initializes an in-memory sqlite3 DB for testing
func setupTestDAO(t *testing.T) (*DAO[TestUser], func()) {
	// Connect to in-memory sqlite3
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to in-memory sqlite: %v", err)
	}

	// Create a test table
	createTableSQL := `CREATE TABLE test_users (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"age" INTEGER
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create DAO instance with the specific type TestUser
	dao := NewDAO[TestUser](db)

	// Teardown function to close the database connection
	teardown := func() {
		db.Close()
	}

	return dao, teardown
}

func TestFullCRUDLifecycle(t *testing.T) {
	dao, teardown := setupTestDAO(t)
	defer teardown()
	ctx := context.Background()

	// 1. BatchInsert
	usersToInsert := []map[string]any{
		{"name": "Peter", "age": 25},
		{"name": "Mary", "age": 30},
		{"name": "John", "age": 35},
	}
	affected, err := dao.BatchInsert(ctx, BatchInsertEndpoint[TestUser]{
		Table: "test_users",
		Rows:  usersToInsert,
	})
	if err != nil {
		t.Fatalf("BatchInsert() error = %v", err)
	}
	if affected != int64(len(usersToInsert)) {
		t.Fatalf("BatchInsert() affected rows = %d, want %d", affected, len(usersToInsert))
	}

	// 2. Select
	var selectedUsers []TestUser
	err = dao.Select(ctx, SelectEndPoint[TestUser]{
		Model:  &selectedUsers,
		Table:  "test_users",
		Fields: []string{"id", "name", "age"},
		Appends: []string{"ORDER BY id ASC"},
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if len(selectedUsers) != 3 {
		t.Fatalf("Select() got %d users, want 3", len(selectedUsers))
	}
	if selectedUsers[0].Name != "Peter" {
		t.Fatalf("Select() user 1 name = %s, want Peter", selectedUsers[0].Name)
	}

	// 3. Update
	newAge := int(40)
	affected, err = dao.Update(ctx, UpdateEndPoint[TestUser]{
		Table: "test_users",
		Rows:  map[string]any{"age": newAge},
		Conditions: map[string]any{"name = ": "John"},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if affected != 1 {
		t.Fatalf("Update() affected rows = %d, want 1", affected)
	}
	var updatedUser TestUser
	dao.Get(ctx, GetEndPoint[TestUser]{Model: &updatedUser, Table: "test_users", Conditions: map[string]any{"id = ": 3}})
	if updatedUser.Age != newAge {
		t.Fatalf("Updated user age = %d, want %d", updatedUser.Age, newAge)
	}

	// 4. Paginate
	var paginatedUsers []TestUser
	total, err := dao.Paginate(ctx, PageEndPoint[TestUser]{
		Model:     &paginatedUsers,
		Table:     "test_users",
		PageNo:    2,
		PageSize:  1,
		SortField: "id",
		SortOrder: "ASC",
	})
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("Paginate() total = %d, want 3", total)
	}
	if len(paginatedUsers) != 1 {
		t.Fatalf("Paginate() got %d users, want 1", len(paginatedUsers))
	}
	if paginatedUsers[0].Name != "Mary" {
		t.Fatalf("Paginate() user name = %s, want Mary", paginatedUsers[0].Name)
	}

	// 5. Delete
	affected, err = dao.Delete(ctx, DeleteEndPoint[TestUser]{
		Table:      "test_users",
		Conditions: map[string]any{"id = ": 1},
	})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if affected != 1 {
		t.Fatalf("Delete() affected rows = %d, want 1", affected)
	}
	var deletedUser TestUser
	err = dao.Get(ctx, GetEndPoint[TestUser]{Model: &deletedUser, Table: "test_users", Conditions: map[string]any{"id = ": 1}})
	if err != sql.ErrNoRows {
		t.Fatalf("Expected sql.ErrNoRows after delete, but got err = %v", err)
	}
}

func TestTransaction(t *testing.T) {
	dao, teardown := setupTestDAO(t)
	defer teardown()
	ctx := context.Background()

	// 1. Test Commit
	txDao, err := dao.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}
	_, err = txDao.Insert(ctx, InsertEndpoint[TestUser]{Table: "test_users", Rows: map[string]any{"name": "CommitUser", "age": 50}})
	if err != nil {
		txDao.Rollback()
		t.Fatalf("tx.Insert() error = %v", err)
	}
	if err = txDao.Commit(); err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}

	var committedUser TestUser
	err = dao.Get(ctx, GetEndPoint[TestUser]{Model: &committedUser, Table: "test_users", Conditions: map[string]any{"name = ": "CommitUser"}})
	if err != nil {
		t.Fatalf("Expected to get committed user, but got err = %v", err)
	}
	if committedUser.Name != "CommitUser" {
		t.Fatalf("Got committed user name = %s, want CommitUser", committedUser.Name)
	}

	// 2. Test Rollback
	txDao, err = dao.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}
	_, err = txDao.Insert(ctx, InsertEndpoint[TestUser]{Table: "test_users", Rows: map[string]any{"name": "RollbackUser", "age": 60}})
	if err != nil {
		txDao.Rollback()
		t.Fatalf("tx.Insert() error = %v", err)
	}
	if err = txDao.Rollback(); err != nil {
		t.Fatalf("tx.Rollback() error = %v", err)
	}

	var rollbackUser TestUser
	err = dao.Get(ctx, GetEndPoint[TestUser]{Model: &rollbackUser, Table: "test_users", Conditions: map[string]any{"name = ": "RollbackUser"}})
	if err != sql.ErrNoRows {
		t.Fatalf("Expected sql.ErrNoRows for rolled back user, but got err = %v", err)
	}
}