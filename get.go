package db_dao

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
)

func (s GetEndPoint[T]) point2Sql() (string, []any, error) {
	var (
		query           string
		fieldsQuery     string
		tableQuery      string
		conditionsQuery string
		conditionsArgs  []any
		appendsQuery    string
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
	if appendsQuery, err = s.appends2string(); err != nil {
		return query, conditionsArgs, errors.New("appends transfer failed")
	}
	query = fmt.Sprintf("SELECT %v FROM %v %v %v", fieldsQuery, tableQuery, conditionsQuery, appendsQuery)

	return query, conditionsArgs, err
}

func (s GetEndPoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return "", errors.New("empty table")
	}
	return s.Table, nil
}

func (s GetEndPoint[T]) fields2string() (string, error) {
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

func (s GetEndPoint[T]) conditions2string() (string, []any, error) {
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

func (s GetEndPoint[T]) appends2string() (string, error) {
	var query string

	if len(s.Appends) > 0 {
		query = strings.Join(s.Appends, " ")
	}
	return query, nil
}
