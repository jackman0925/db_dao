package db_dao

import (
	"fmt"
	"strings"
)

// for count
func (s PageEndPoint[T]) point2Sql() (string, []any, error) {
	tableQuery, err := buildTableClause(s.Table)
	if err != nil {
		return "", nil, err
	}

	conditionsQuery, conditionsArgs, err := buildWhereClause(s.Conditions)
	if err != nil {
		return "", nil, err
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %v %v", tableQuery, conditionsQuery)

	return query, conditionsArgs, nil
}

// for select
func (s PageEndPoint[T]) point2pageSql() (string, []any, error) {
	fieldsQuery := buildFieldsClause(s.Fields)

	tableQuery, err := buildTableClause(s.Table)
	if err != nil {
		return "", nil, err
	}

	conditionsQuery, conditionsArgs, err := buildWhereClause(s.Conditions)
	if err != nil {
		return "", nil, err
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("SELECT %v FROM %v", fieldsQuery, tableQuery))

	if conditionsQuery != "" {
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(conditionsQuery)
	}

	if s.SortField != "" {
		order := "ASC" // 默认为 ASC
		if strings.ToUpper(s.SortOrder) == "DESC" {
			order = "DESC"
		}
		// 注意：为了最大程度的安全，SortField 应根据业务逻辑进行白名单验证。
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", s.SortField, order))
	}

	queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", s.PageSize, (s.PageNo-1)*s.PageSize))

	return queryBuilder.String(), conditionsArgs, nil
}