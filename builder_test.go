package db_dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- builder_test.go: Tests for the SQL builder functions ---

func TestBuildTableClause(t *testing.T) {
	t.Run("valid table", func(t *testing.T) {
		result, err := buildTableClause("users")
		assert.NoError(t, err)
		assert.Equal(t, "users", result)
	})

	t.Run("empty table", func(t *testing.T) {
		_, err := buildTableClause("")
		assert.Error(t, err)
		assert.Equal(t, "empty table", err.Error())
	})
}

func TestBuildFieldsClause(t *testing.T) {
	t.Run("empty fields returns *", func(t *testing.T) {
		assert.Equal(t, "*", buildFieldsClause(nil))
		assert.Equal(t, "*", buildFieldsClause([]string{}))
	})

	t.Run("single wildcard returns *", func(t *testing.T) {
		assert.Equal(t, "*", buildFieldsClause([]string{"*"}))
	})

	t.Run("specific fields", func(t *testing.T) {
		assert.Equal(t, "id,name,age", buildFieldsClause([]string{"id", "name", "age"}))
	})
}

func TestBuildWhereClause(t *testing.T) {
	t.Run("empty conditions", func(t *testing.T) {
		query, args, err := buildWhereClause(nil)
		assert.NoError(t, err)
		assert.Equal(t, "", query)
		assert.Nil(t, args)
	})

	t.Run("single condition", func(t *testing.T) {
		query, args, err := buildWhereClause(map[string]any{"id = ": 1})
		assert.NoError(t, err)
		assert.Equal(t, "WHERE (id = ?)", query)
		assert.Equal(t, []any{1}, args)
	})
}

func TestBuildConditions_InClause(t *testing.T) {
	t.Run("IN with slice generates correct SQL", func(t *testing.T) {
		query, args, err := buildConditions(map[string]any{
			"id": []int{1, 2, 3},
		})
		require.NoError(t, err)
		// Should produce: (id IN (?, ?, ?))
		assert.Contains(t, query, "id IN")
		assert.Contains(t, query, "?, ?, ?")
		assert.NotContains(t, query, "IN IN") // The old bug was generating double IN
		assert.Equal(t, []any{1, 2, 3}, args)
	})

	t.Run("IN with string slice", func(t *testing.T) {
		query, args, err := buildConditions(map[string]any{
			"name": []string{"Alice", "Bob"},
		})
		require.NoError(t, err)
		assert.Contains(t, query, "name IN")
		assert.Contains(t, query, "?, ?")
		assert.Equal(t, []any{"Alice", "Bob"}, args)
	})
}

func TestBuildConditions_NilValue(t *testing.T) {
	t.Run("nil value does not panic", func(t *testing.T) {
		// This used to panic because reflect.ValueOf(nil).Kind() panics
		assert.NotPanics(t, func() {
			query, _, err := buildConditions(map[string]any{
				"deleted_at IS ": nil,
			})
			assert.NoError(t, err)
			assert.Contains(t, query, "deleted_at IS  NULL")
		})
	})
}

func TestBuildSetClauseForUpdate(t *testing.T) {
	t.Run("empty rows", func(t *testing.T) {
		_, _, err := buildSetClauseForUpdate(nil)
		assert.Error(t, err)
		assert.Equal(t, "empty rows for update", err.Error())
	})

	t.Run("single row", func(t *testing.T) {
		query, args, err := buildSetClauseForUpdate(map[string]any{"name": "Alice"})
		assert.NoError(t, err)
		assert.Equal(t, "name = ?", query)
		assert.Equal(t, []any{"Alice"}, args)
	})
}

func TestBuildAppendsClause(t *testing.T) {
	t.Run("empty appends", func(t *testing.T) {
		assert.Equal(t, "", buildAppendsClause(nil))
	})

	t.Run("with appends", func(t *testing.T) {
		assert.Equal(t, "ORDER BY id ASC LIMIT 10", buildAppendsClause([]string{"ORDER BY id ASC", "LIMIT 10"}))
	})
}
