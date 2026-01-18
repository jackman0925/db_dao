package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // Standard library bindings for pgx
	"github.com/jackman0925/db_dao"
	"github.com/jmoiron/sqlx"
)

// User represents the user model.
// Note: For Postgres, we often use `db` tags.
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func main() {
	// 1. Get database connection string from environment variable.
	// You can run this example with: PG_DSN="postgres://user:password@localhost:5432/dbname?sslmode=disable" go run main.go
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		fmt.Println("Please set PG_DSN environment variable to run this example.")
		fmt.Println("Example: export PG_DSN=\"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable\"")
		return
	}

	// 2. Connect to the database using the "pgx" driver.
	// We use sqlx.Connect which wraps sql.Open and also pings the database.
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("--- Connected to Postgres using pgx driver ---")

	// 3. Create the table for demonstration (if not exists).
	schema := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT,
		age INTEGER
	)`
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 4. Initialize the DAO with the pgx connection.
	// The DAO will automatically handle query rebinding ($1, $2, etc.) because sqlx detects the driver.
	userDAO := db_dao.NewDAO[User](db)
	ctx := context.Background()

	// 5. Insert a user
	fmt.Println("\n--- Inserting User ---")
	// Note: For Postgres SERIAL, we usually let DB handle ID, but InsertEndpoint supports returning ID if needed.
	// However, the base Insert method in dao.go uses Exec, which returns LastInsertId.
	// IMPORTANT: lib/pq and pgx do NOT support LastInsertId() for standard Exec calls.
	// To get the ID back in Postgres, we usually need "RETURNING id" and QueryRow.
	//
	// Current DAO implementation uses `result.RowsAffected()` which works, but `result.LastInsertId()` will fail on PG.
	// Let's see if we can use it.

	insertData := db_dao.InsertEndpoint[User]{
		Table: "users",
		Rows:  map[string]any{"name": "PgUser", "age": 25},
	}

	rowsAffected, err := userDAO.Insert(ctx, insertData)
	if err != nil {
		log.Fatalf("Insert failed: %v", err)
	}
	fmt.Printf("Rows affected: %d\n", rowsAffected)

	// 6. Select users
	fmt.Println("\n--- Selecting Users ---")
	var users []User
	selectData := db_dao.SelectEndPoint[User]{
		Table:      "users",
		Model:      &users,
		Conditions: map[string]any{"name =": "PgUser"},
	}

	err = userDAO.Select(ctx, selectData)
	if err != nil {
		log.Fatalf("Select failed: %v", err)
	}

	for _, u := range users {
		fmt.Printf("User: ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
	}

	// Cleanup (Optional)
	// _, _ = db.Exec("DROP TABLE users")
}
