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
	// 1. Open a connection to a SQLite database.
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Create the users table.
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
	fmt.Println("---", "Database table 'users' created.", "---")

	// 3. Create a new DAO for the User model.
	// NewDAO returns a DAO instance that satisfies the IDAO interface.
	var userDAO db_dao.IDAO[User] = db_dao.NewDAO[User](db)

	// --- Basic Operations (Non-Transactional) ---
	fmt.Println("\n--- Running basic non-transactional operations... ---")
	runBasicOperations(userDAO)

	// --- Transactional Operations ---
	fmt.Println("\n--- Running operations within a transaction... ---")
	runTransactionalOperations(userDAO)

	// --- Final State Check ---
	fmt.Println("\n--- Final check of the database state...")
	var users []User
	_ = userDAO.Select(context.Background(), db_dao.SelectEndPoint[User]{Model: &users, Table: "users"})
	fmt.Printf("Found %d users in total:\n", len(users))
	for _, u := range users {
		fmt.Printf("- ID: %d, Name: %s, Age: %d\n", u.ID, u.Name, u.Age)
	}
}

func runBasicOperations(userDAO db_dao.IDAO[User]) {
	ctx := context.Background()

	// Insert a new user
	insertEndpoint := db_dao.InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "Alice", "age": 30},
	}
	insertedID, err := userDAO.Insert(ctx, insertEndpoint)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}
	fmt.Printf("Inserted user 'Alice' with ID: %d\n", insertedID)

	// Get a user
	var alice User
	getEndpoint := db_dao.GetEndPoint[User]{
		Table:      "users",
		Model:      &alice,
		Conditions: map[string]any{"name = ": "Alice"},
	}
	err = userDAO.Get(ctx, getEndpoint)
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}
	fmt.Printf("Retrieved user: ID=%d, Name=%s, Age=%d\n", alice.ID, alice.Name, alice.Age)
}

func runTransactionalOperations(userDAO db_dao.IDAO[User]) {
	ctx := context.Background()

	// 1. Begin a transaction
	txDAO, err := userDAO.BeginTx(ctx)
	if err != nil {
		log.Fatalf("failed to begin transaction: %v", err)
	}
	fmt.Println("Transaction started.")

	// Use a deferred function to handle commit or rollback
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic, rolling back transaction...")
			txDAO.Rollback()
			panic(r) // re-panic after rollback
		} else if err != nil {
			fmt.Println("Error occurred, rolling back transaction...")
			txDAO.Rollback()
		} else {
			fmt.Println("Committing transaction...")
			err = txDAO.Commit()
		}
	}()

	// 2. Perform operations within the transaction
	fmt.Println("Inserting 'Bob' within transaction...")
	_, err = txDAO.Insert(ctx, db_dao.InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "Bob", "age": 40},
	})
	if err != nil {
		return // The deferred function will handle the rollback
	}

	fmt.Println("Updating 'Alice' within transaction...")
	_, err = txDAO.Update(ctx, db_dao.UpdateEndPoint[User]{
		Table:      "users",
		Rows:       map[string]any{"age": 31},
		Conditions: map[string]any{"name = ": "Alice"},
	})
	if err != nil {
		return // The deferred function will handle the rollback
	}

	// If we reach here, the deferred function will commit the transaction.
}