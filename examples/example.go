
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackman0925/db_dao"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// User is a sample struct representing a user in the database.
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func main() {
	// Open a connection to a SQLite database.
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create the users table.
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		age INTEGER
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// Create a new DAO for the User model.
	userDAO := db_dao.NewDAO[User](db)

	// --- Insert a new user ---
	insertEndpoint := db_dao.InsertEndpoint[User]{
		Table: "users",
		Rows: []db_dao.Row{
			{Column: "name", Value: "Alice"},
			{Column: "age", Value: 25},
		},
	}
	inserted, err := userDAO.Insert(context.Background(), insertEndpoint)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}
	fmt.Printf("Inserted %d user(s)
", inserted)

	// --- Get a user ---
	var alice User
	getEndpoint := db_dao.GetEndPoint[User]{
		Table: "users",
		Model: &alice,
		Conditions: []db_dao.Condition{
			{Column: "name", Operator: "=", Value: "Alice"},
		},
	}
	err = userDAO.Get(context.Background(), getEndpoint)
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}
	fmt.Printf("Retrieved user: ID=%d, Name=%s, Age=%d
", alice.ID, alice.Name, alice.Age)

	// --- Select multiple users ---
	var users []User
	selectEndpoint := db_dao.SelectEndPoint[User]{
		Table: "users",
		Model: &users,
	}
	err = userDAO.Select(context.Background(), selectEndpoint)
	if err != nil {
		log.Fatalf("failed to select users: %v", err)
	}
	fmt.Printf("Selected %d user(s)
", len(users))
	for _, u := range users {
		fmt.Printf("- User: ID=%d, Name=%s, Age=%d
", u.ID, u.Name, u.Age)
	}
}
