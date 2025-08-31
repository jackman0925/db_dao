package db_dao

import (
	"errors"
	"fmt"
	"strings"
)

func (s UpdateEndPoint[T]) point2Sql() (string, []any, []any, error) {
	tableQuery, err := buildTableClause(s.Table)
	if err != nil {
		return "", nil, nil, err
	}

	rowsQuery, rowsArgs, err := buildSetClauseForUpdate(s.Rows)
	if err != nil {
		return "", nil, nil, err
	}

	// For safety, UPDATE must have conditions
	if len(s.Conditions) == 0 {
		return "", nil, nil, errors.New("empty conditions for update")
	}

	conditionsQuery, conditionsArgs, err := buildWhereClause(s.Conditions)
	if err != nil {
		return "", nil, nil, err
	}

	appendsQuery := buildAppendsClause(s.Appends)

	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("UPDATE %v SET %v", tableQuery, rowsQuery))

	if conditionsQuery != "" {
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(conditionsQuery)
	}

	if appendsQuery != "" {
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(appendsQuery)
	}

	return queryBuilder.String(), rowsArgs, conditionsArgs, nil
}