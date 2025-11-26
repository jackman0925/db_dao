package db_dao

import (
	"errors"
	"fmt"
	"reflect"
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
	for k, v := range conditions {
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

		// 注意: 此处保留了原始实现中的逻辑。
		// 对于 value 是切片的情况，原始逻辑可能存在问题，因为它会生成像 `(field IN IN (?,?))` 这样的SQL。
		// 在后续步骤中可以修复此问题。
		if reflect.ValueOf(v).Kind() == reflect.Slice {
			var (
				inQuery string
				inArgs  []any
			)
			// 原始逻辑
			inQuery, inArgs, _ = sqlx.In(" IN (?)", v)

			k = fmt.Sprintf("(%v%v)", k, inQuery) // 移除原始逻辑中多余的空格
			args = append(args, inArgs...)
		} else {
			k = fmt.Sprintf("(%v?)", k) // 移除原始逻辑中多余的空格
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
	for k, v := range rows {
		prepareRows = append(prepareRows, fmt.Sprintf("%v = ?", k))
		args = append(args, v)
	}
	return strings.Join(prepareRows, ","), args, nil
}
