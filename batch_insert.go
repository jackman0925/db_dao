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
	// Get fields from the first row
	for k := range s.Rows[0] {
		prepareFields = append(prepareFields, k)
	}

	for _, row := range s.Rows {
		var rowValues []string
		for _, fieldName := range prepareFields {
			rowValues = append(rowValues, "?")
			args = append(args, row[fieldName])
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
