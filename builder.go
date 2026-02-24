package db_dao

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// buildTableClause 构建 FROM 子句
func buildTableClause(table string) (string, error) {
	if table == "" {
		return "", errors.New("empty table")
	}
	return table, nil
}

// buildFieldsClause 构建 SELECT 的字段部分
func buildFieldsClause(fields []string) string {
	if len(fields) == 0 || (len(fields) == 1 && fields[0] == "*") {
		return "*"
	}
	return strings.Join(fields, ",")
}

// buildWhereClause 从 conditions map 构建 WHERE 子句
func buildWhereClause(conditions map[string]any) (string, []any, error) {
	query, args, err := buildConditions(conditions)
	if err != nil {
		return "", nil, err
	}
	if query == "" {
		return "", nil, nil
	}
	return fmt.Sprintf("WHERE %v", query), args, nil
}

func buildConditions(conditions map[string]any) (string, []any, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}

	var (
		prepareConditions []string
		args              []any
	)
	for _, k := range sortedKeys(conditions) {
		v := conditions[k]
		if orConds, ok := v.(Or); ok {
			var orParts []string
			for _, subCond := range orConds {
				subQuery, subArgs, err := buildConditions(subCond)
				if err != nil {
					return "", nil, err
				}
				if subQuery != "" {
					orParts = append(orParts, fmt.Sprintf("(%s)", subQuery))
					args = append(args, subArgs...)
				}
			}
			if len(orParts) > 0 {
				prepareConditions = append(prepareConditions, fmt.Sprintf("(%s)", strings.Join(orParts, " OR ")))
			}
			continue
		}

		// Handle nil values — cannot use reflect on nil
		if v == nil {
			prepareConditions = append(prepareConditions, fmt.Sprintf("(%v NULL)", k))
			continue
		}

		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			var (
				inQuery string
				inArgs  []any
				err     error
			)
			inQuery, inArgs, err = sqlx.In("(?)", v)
			if err != nil {
				return "", nil, err
			}

			k = fmt.Sprintf("(%v IN %v)", k, inQuery)
			args = append(args, inArgs...)
		} else {
			k = fmt.Sprintf("(%v?)", k)
			args = append(args, v)
		}
		prepareConditions = append(prepareConditions, k)
	}
	return strings.Join(prepareConditions, " AND "), args, nil
}

// buildAppendsClause 构建追加的SQL语句 (如 ORDER BY, GROUP BY)
func buildAppendsClause(appends []string) string {
	if len(appends) > 0 {
		return strings.Join(appends, " ")
	}
	return ""
}

// buildSetClauseForUpdate 构建 UPDATE 的 SET 子句
func buildSetClauseForUpdate(rows map[string]any) (string, []any, error) {
	if len(rows) == 0 {
		return "", nil, errors.New("empty rows for update")
	}
	var (
		prepareRows []string
		args        []any
	)
	for _, k := range sortedKeys(rows) {
		v := rows[k]
		prepareRows = append(prepareRows, fmt.Sprintf("%v = ?", k))
		args = append(args, v)
	}
	return strings.Join(prepareRows, ","), args, nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
