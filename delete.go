package db_dao

import (
	"errors"
	"fmt"
)

func (s DeleteEndPoint[T]) point2Sql() (string, []any, error) {
	tableQuery, err := buildTableClause(s.Table)
	if err != nil {
		return "", nil, err
	}

	// For safety, DELETE must have conditions
	if len(s.Conditions) == 0 {
		return "", nil, errors.New("empty conditions for delete")
	}

	conditionsQuery, conditionsArgs, err := buildWhereClause(s.Conditions)
	if err != nil {
		return "", nil, err
	}

	query := fmt.Sprintf("DELETE FROM %v %v", tableQuery, conditionsQuery)

	return query, conditionsArgs, nil
}