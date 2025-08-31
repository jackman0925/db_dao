package db_dao

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"strings"
)

func (s DeleteEndPoint[T]) point2Sql() (string, []any, error) {
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
	query = fmt.Sprintf("DELETE FROM %v %v", tableQuery, conditionsQuery)

	return query, conditionsArgs, err
}

func (s DeleteEndPoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return "", errors.New("empty table")
	}
	return s.Table, nil
}

func (s DeleteEndPoint[T]) conditions2string() (string, []any, error) {
	var (
		query             string
		prepareConditions []string
		args              []any
		err               error
	)
	if len(s.Conditions) == 0 {
		return query, args, errors.New("empty conditions for delete") // Safety break
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
