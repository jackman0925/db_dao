package db_dao

import (
	"errors"
	"fmt"
	"strings"
)

func (s InsertEndpoint[T]) point2Sql() (string, []any, error) {
	var (
		query       string
		tableQuery  string
		fieldsQuery string
		valuesQuery string
		rowsArgs    []any
		err         error
	)
	if tableQuery, err = s.table2string(); err != nil {
		return query, rowsArgs, errors.New("table transfer failed")
	}
	if fieldsQuery, valuesQuery, rowsArgs, err = s.rows2sql(); err != nil {
		return query, rowsArgs, errors.New("rows transfer failed")
	}
	query = fmt.Sprintf("INSERT INTO %v %v VALUES %v", tableQuery, fieldsQuery, valuesQuery)

	return query, rowsArgs, err
}

func (s InsertEndpoint[T]) rows2sql() (string, string, []any, error) {
	if len(s.Rows) == 0 {
		return "", "", nil, errors.New("empty rows")
	}
	var (
		fieldsQuery   string
		valuesQuery   string
		prepareFields []string
		prepareRows   []string
		args          []any
	)
	for k, v := range s.Rows {
		prepareFields = append(prepareFields, k)
		prepareRows = append(prepareRows, "?")
		args = append(args, v)
	}
	fieldsQuery = fmt.Sprintf("(%v)", strings.Join(prepareFields, ","))
	valuesQuery = fmt.Sprintf("(%v)", strings.Join(prepareRows, ","))

	return fieldsQuery, valuesQuery, args, nil
}

func (s InsertEndpoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return "", errors.New("empty table")
	}
	return s.Table, nil
}
