package db_dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"testing"
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
	);
`
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

func TestInsertAndGet(t *testing.T) {
	dao, teardown := setupTestDAO(t)
	defer teardown()

	// 1. Test Insert
	userName := "John Doe"
	userAge := 30
	affectedRows, err := dao.Insert(InsertEndpoint[TestUser]{
		Table: "test_users",
		Rows: map[string]any{
			"name": userName,
			"age":  userAge,
		},
	})

	if err != nil {
		t.Fatalf("Insert() returned an unexpected error: %v", err)
	}
	if affectedRows != 1 {
		t.Fatalf("Expected 1 row to be affected, but got %d", affectedRows)
	}

	// 2. Test Get
	var retrievedUser TestUser
	err = dao.Get(GetEndPoint[TestUser]{
		Model: &retrievedUser,
		Table: "test_users",
		Fields: []string{"id", "name", "age"},
		Conditions: map[string]any{
			"id = ": 1,
		},
	})

	if err != nil {
		t.Fatalf("Get() returned an unexpected error: %v", err)
	}

	// 3. Verify the result
	expectedUser := TestUser{ID: 1, Name: userName, Age: userAge}
	if !reflect.DeepEqual(retrievedUser, expectedUser) {
		t.Fatalf("Retrieved user does not match expected user.\nExpected: %+v\nGot:      %+v", expectedUser, retrievedUser)
	}

	fmt.Println("TestInsertAndGet PASSED")
}