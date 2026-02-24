package db_dao

import (
	"errors"
	"fmt"
	"strings"
)

func (s BatchInsertEndpoint[T]) point2Sql() (string, []any, error) {
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

func (s BatchInsertEndpoint[T]) rows2sql() (string, string, []any, error) {
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
	prepareFields = sortedKeys(s.Rows[0])

	for _, row := range s.Rows {
		if len(row) != len(prepareFields) {
			return "", "", nil, errors.New("inconsistent row fields")
		}
		var rowValues []string
		for _, fieldName := range prepareFields {
			v, ok := row[fieldName]
			if !ok {
				return "", "", nil, errors.New("inconsistent row fields")
			}
			rowValues = append(rowValues, "?")
			args = append(args, v)
		}
		prepareRows = append(prepareRows, "("+strings.Join(rowValues, ",")+")")
	}
	fieldsQuery = fmt.Sprintf("(%v)", strings.Join(prepareFields, ","))
	valuesQuery = strings.Join(prepareRows, ",")

	return fieldsQuery, valuesQuery, args, nil
}

func (s BatchInsertEndpoint[T]) table2string() (string, error) {
	if s.Table == "" {
		return "", errors.New("empty table")
	}
	return s.Table, nil
}
