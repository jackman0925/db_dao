package db_dao

import (
	"fmt"
	"strings"
)

func (s SelectEndPoint[T]) point2Sql() (string, []any, error) {
	fieldsQuery := buildFieldsClause(s.Fields)

	tableQuery, err := buildTableClause(s.Table)
	if err != nil {
		return "", nil, err
	}

	conditionsQuery, conditionsArgs, err := buildWhereClause(s.Conditions)
	if err != nil {
		return "", nil, err
	}

	appendsQuery := buildAppendsClause(s.Appends)

	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("SELECT %v FROM %v", fieldsQuery, tableQuery))

	if conditionsQuery != "" {
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(conditionsQuery)
	}

	if appendsQuery != "" {
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(appendsQuery)
	}

	return queryBuilder.String(), conditionsArgs, nil
}