package db_dao

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
)

// for count
func (s PageEndPoint[T]) point2Sql() (string, []any, error) {
	var (
		query           string
		tableQuery      string
		conditionsQuery string
		conditionsArgs  []any
		err             error
	)
	if tableQuery, err = s.table2string(); err != nil {
		return query, conditionsArgs, errors.New("table transfer failed")
	}
	if conditionsQuery, conditionsArgs, err = s.conditions2string(); err != nil {
		return query, conditionsArgs, errors.New("condition transfer failed")
	}

	query = fmt.Sprintf("SELECT COUNT(*) FROM %v %v", tableQuery, conditionsQuery)

	return query, conditionsArgs, err
}

// for select
func (s PageEndPoint[T]) point2pageSql() (string, []any, error) {
	var (
		query           string
		tableQuery      string
		conditionsQuery string
		conditionsArgs  []any
		fieldsQuery     string
		err             error
	)
	if fieldsQuery, err = s.fields2string(); err != nil {
		return query, conditionsArgs, errors.New("fields transfer failed")
	}
	if tableQuery, err = s.table2string(); err != nil {
		return query, conditionsArgs, errors.New("table transfer failed")
	}
	if conditionsQuery, conditionsArgs, err = s.conditions2string(); err != nil {
		return query, conditionsArgs, errors.New("condition transfer failed")
	}
	query = fmt.Sprintf("SELECT %v FROM %v %v", fieldsQuery, tableQuery, conditionsQuery)

	if s.SortBy != "" {
		query += " ORDER BY " + s.SortBy
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", s.PageSize, (s.PageNo-1)*s.PageSize)

	return query, conditionsArgs, err
}

func (s PageEndPoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return "", errors.New("empty table")
	}
	return s.Table, nil
}

func (s PageEndPoint[T]) fields2string() (string, error) {
	var query string

	if len(s.Fields) == 0 {
		return "*", nil
	}
	if s.Fields[0] == "*" {
		query = "*"
	} else {
		query = strings.Join(s.Fields, ",")
	}
	return query, nil
}

func (s PageEndPoint[T]) conditions2string() (string, []any, error) {
	var (
		query             string
		prepareConditions []string
		args              []any
		err               error
	)
	if len(s.Conditions) == 0 {
		return query, args, err
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
