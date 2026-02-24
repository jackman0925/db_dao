package db_dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- endpoint_test.go: Tests for each endpoint's point2Sql() ---

func TestGetEndPoint_point2Sql(t *testing.T) {
	t.Run("basic get", func(t *testing.T) {
		var u struct{}
		ep := GetEndPoint[struct{}]{
			Model:      &u,
			Table:      "users",
			Conditions: map[string]any{"id = ": 1},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE (id = ?)", query)
		assert.Equal(t, []any{1}, args)
	})

	t.Run("with fields and appends", func(t *testing.T) {
		var u struct{}
		ep := GetEndPoint[struct{}]{
			Model:      &u,
			Table:      "users",
			Fields:     []string{"id", "name"},
			Conditions: map[string]any{"id = ": 1},
			Appends:    []string{"LIMIT 1"},
		}
		query, _, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "SELECT id,name FROM users")
		assert.Contains(t, query, "LIMIT 1")
	})

	t.Run("empty table", func(t *testing.T) {
		var u struct{}
		ep := GetEndPoint[struct{}]{Model: &u, Table: ""}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})
}

func TestSelectEndPoint_point2Sql(t *testing.T) {
	t.Run("basic select without conditions", func(t *testing.T) {
		var users []struct{}
		ep := SelectEndPoint[struct{}]{
			Model: &users,
			Table: "users",
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users", query)
		assert.Nil(t, args)
	})
}

func TestInsertEndpoint_point2Sql(t *testing.T) {
	t.Run("empty table", func(t *testing.T) {
		ep := InsertEndpoint[struct{}]{Table: "", Rows: map[string]any{"a": 1}}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("empty rows", func(t *testing.T) {
		ep := InsertEndpoint[struct{}]{Table: "users", Rows: nil}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("valid insert", func(t *testing.T) {
		ep := InsertEndpoint[struct{}]{
			Table: "users",
			Rows:  map[string]any{"name": "Alice"},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "INSERT INTO users")
		assert.Contains(t, query, "name")
		assert.Contains(t, args, "Alice")
	})

	t.Run("deterministic field order", func(t *testing.T) {
		ep := InsertEndpoint[struct{}]{
			Table: "users",
			Rows:  map[string]any{"name": "Alice", "age": 30},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "INSERT INTO users (age,name) VALUES (?,?)", query)
		assert.Equal(t, []any{30, "Alice"}, args)
	})
}

func TestBatchInsertEndpoint_point2Sql(t *testing.T) {
	t.Run("empty table", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{Table: "", Rows: []map[string]any{{"a": 1}}}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("empty rows", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{Table: "users", Rows: nil}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("valid batch insert with single row", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{
			Table: "users",
			Rows: []map[string]any{
				{"name": "Alice", "age": 30},
			},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "INSERT INTO users")
		assert.Len(t, args, 2)
	})

	t.Run("valid batch insert with multiple rows", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{
			Table: "users",
			Rows: []map[string]any{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 40},
			},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "INSERT INTO users")
		assert.Len(t, args, 4)
		// Should have two value groups
		assert.Contains(t, query, "VALUES")
	})

	t.Run("deterministic field order", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{
			Table: "users",
			Rows: []map[string]any{
				{"name": "Alice", "age": 30},
				{"age": 40, "name": "Bob"},
			},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "INSERT INTO users (age,name) VALUES (?,?),(?,?)", query)
		assert.Equal(t, []any{30, "Alice", 40, "Bob"}, args)
	})

	t.Run("inconsistent row fields returns error", func(t *testing.T) {
		ep := BatchInsertEndpoint[struct{}]{
			Table: "users",
			Rows: []map[string]any{
				{"name": "Alice", "age": 30},
				{"name": "Bob"},
			},
		}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
		assert.Equal(t, "rows transfer failed", err.Error())
	})
}

func TestUpdateEndPoint_point2Sql(t *testing.T) {
	t.Run("empty conditions", func(t *testing.T) {
		ep := UpdateEndPoint[struct{}]{
			Table: "users",
			Rows:  map[string]any{"age": 31},
		}
		_, _, _, err := ep.point2Sql()
		assert.Error(t, err)
		assert.Equal(t, "empty conditions for update", err.Error())
	})

	t.Run("empty rows", func(t *testing.T) {
		ep := UpdateEndPoint[struct{}]{
			Table:      "users",
			Conditions: map[string]any{"id = ": 1},
		}
		_, _, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("valid update", func(t *testing.T) {
		ep := UpdateEndPoint[struct{}]{
			Table:      "users",
			Rows:       map[string]any{"age": 31},
			Conditions: map[string]any{"id = ": 1},
		}
		query, rowsArgs, conditionsArgs, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "UPDATE users SET age = ?")
		assert.Contains(t, query, "WHERE (id = ?)")
		assert.Equal(t, []any{31}, rowsArgs)
		assert.Equal(t, []any{1}, conditionsArgs)
	})
}

func TestDeleteEndPoint_point2Sql(t *testing.T) {
	t.Run("empty conditions", func(t *testing.T) {
		ep := DeleteEndPoint[struct{}]{Table: "users"}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
		assert.Equal(t, "empty conditions for delete", err.Error())
	})

	t.Run("empty table", func(t *testing.T) {
		ep := DeleteEndPoint[struct{}]{
			Table:      "",
			Conditions: map[string]any{"id = ": 1},
		}
		_, _, err := ep.point2Sql()
		assert.Error(t, err)
	})

	t.Run("valid delete", func(t *testing.T) {
		ep := DeleteEndPoint[struct{}]{
			Table:      "users",
			Conditions: map[string]any{"id = ": 1},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "DELETE FROM users WHERE (id = ?)", query)
		assert.Equal(t, []any{1}, args)
	})
}

func TestPageEndPoint_point2Sql(t *testing.T) {
	t.Run("count query", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model: &users,
			Table: "users",
		}
		query, _, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Equal(t, "SELECT COUNT(*) FROM users ", query)
	})

	t.Run("count query with conditions", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:      &users,
			Table:      "users",
			Conditions: map[string]any{"age > ": 18},
		}
		query, args, err := ep.point2Sql()
		require.NoError(t, err)
		assert.Contains(t, query, "SELECT COUNT(*) FROM users WHERE")
		assert.Equal(t, []any{18}, args)
	})
}

func TestPageEndPoint_point2pageSql(t *testing.T) {
	t.Run("basic paging", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:    &users,
			Table:    "users",
			PageNo:   1,
			PageSize: 10,
		}
		query, _, err := ep.point2pageSql()
		require.NoError(t, err)
		assert.Contains(t, query, "SELECT * FROM users")
		assert.Contains(t, query, "LIMIT 10 OFFSET 0")
	})

	t.Run("second page", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:    &users,
			Table:    "users",
			PageNo:   2,
			PageSize: 10,
		}
		query, _, err := ep.point2pageSql()
		require.NoError(t, err)
		assert.Contains(t, query, "LIMIT 10 OFFSET 10")
	})

	t.Run("with sort", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:     &users,
			Table:     "users",
			PageNo:    1,
			PageSize:  10,
			SortField: "created_at",
			SortOrder: "DESC",
		}
		query, _, err := ep.point2pageSql()
		require.NoError(t, err)
		assert.Contains(t, query, "ORDER BY created_at DESC")
	})

	t.Run("sort defaults to ASC", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:     &users,
			Table:     "users",
			PageNo:    1,
			PageSize:  10,
			SortField: "id",
		}
		query, _, err := ep.point2pageSql()
		require.NoError(t, err)
		assert.Contains(t, query, "ORDER BY id ASC")
	})

	t.Run("invalid pageNo = 0", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:    &users,
			Table:    "users",
			PageNo:   0,
			PageSize: 10,
		}
		_, _, err := ep.point2pageSql()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pageNo")
	})

	t.Run("invalid pageNo = -1", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:    &users,
			Table:    "users",
			PageNo:   -1,
			PageSize: 10,
		}
		_, _, err := ep.point2pageSql()
		assert.Error(t, err)
	})

	t.Run("invalid pageSize = 0", func(t *testing.T) {
		var users []struct{}
		ep := PageEndPoint[struct{}]{
			Model:    &users,
			Table:    "users",
			PageNo:   1,
			PageSize: 0,
		}
		_, _, err := ep.point2pageSql()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pageSize")
	})
}
