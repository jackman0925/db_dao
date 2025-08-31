package db_dao

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
)

func (s UpdateEndPoint[T]) point2Sql() (string, []any, []any, error) {
	var (
		query           string
		tableQuery      string
		rowsQuery       string
		rowsArgs        []any
		conditionsQuery string
		conditionsArgs  []any
		appendsQuery    string
		err             error
	)
	if tableQuery, err = s.table2string(); err != nil {
		return query, rowsArgs, conditionsArgs, errors.New("table transfer failed")
	}
	if rowsQuery, rowsArgs, err = s.rows2string(); err != nil {
		return query, rowsArgs, conditionsArgs, errors.New("rows transfer failed")
	}
	if conditionsQuery, conditionsArgs, err = s.conditions2string(); err != nil {
		return query, rowsArgs, conditionsArgs, errors.New("condition transfer failed")
	}
	if appendsQuery, err = s.appends2string(); err != nil {
		return query, rowsArgs, conditionsArgs, errors.New("appends transfer failed")
	}
	query = fmt.Sprintf("UPDATE %v SET %v %v %v", tableQuery, rowsQuery, conditionsQuery, appendsQuery)

	return query, rowsArgs, conditionsArgs, err
}

func (s UpdateEndPoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return s.Table, errors.New("empty table")
	}
	return s.Table, nil
}

func (s UpdateEndPoint[T]) rows2string() (string, []any, error) {
	if len(s.Rows) == 0 {
		return "", nil, errors.New("empty rows")
	}
	var (
		query       string
		prepareRows []string
		args        []any
	)
	for k, v := range s.Rows {
		prepareRows = append(prepareRows, fmt.Sprintf("%v = ?", k))
		args = append(args, v)
	}
	query = strings.Join(prepareRows, ",")
	return query, args, nil
}

func (s UpdateEndPoint[T]) conditions2string() (string, []any, error) {
	var (
		query             string
		prepareConditions []string
		args              []any
		err               error
	)
	if len(s.Conditions) == 0 {
		return query, args, errors.New("empty conditions for update") // Safety break
	}
	for k, v := range s.Conditions {
		if reflect.ValueOf(v).Kind() == reflect.Slice {
			var (
				inQuery string
				inArgs  []any
			)
			inQuery, inArgs, _ = sqlx.In(" IN (?)", v)

			k = fmt.Sprintf("(%v %v)", k, inQuery)
			args = append(args, inArgs...)
		} else {
			k = fmt.Sprintf("(%v %v)", k, "?")
			args = append(args, v)
		}
		prepareConditions = append(prepareConditions, k)
	}
	query = fmt.Sprintf("WHERE %v", strings.Join(prepareConditions, " AND "))
	return query, args, nil
}

func (s UpdateEndPoint[T]) appends2string() (string, error) {
	var query string

	if len(s.Appends) > 0 {
		query = strings.Join(s.Appends, " ")
	}
	return query, nil
}
