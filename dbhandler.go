package main

import (
	"database/sql"
	"log"
	"strings"
)

const (
	DB_DRIVER = "sqlite3"
	DB_PATH   = "./database/sample-lobbo.db"

	// DB CRUD
	CRUD_MAX = 6
	INS      = iota
	SEL
	UPD
	DEL
)

var (
	dbFail func(...interface{}) = log.Fatalln

	// SQL command templates
	sqlCRUD = [][]string{
		{"INSERT INTO ", "", " VALUES ", ""},
		{"SELECT (", "", ") FROM ", "", " WHERE ", ""},
		{"UPDATE ", "", " SET ", "", " WHERE ", ""},
		{"DELETE FROM ", "", " WHERE ", ""},
	}
)

type DB struct {
	*sql.DB
}

func ConnectDB() (*DB, error) {
	db, err := sql.Open(DB_DRIVER, DB_PATH)

	return &DB{db}, err
}

// func (db *DB) Insert(table string, cols []string, vals []string) error {
// 	// inStr := prepString("insert", table, cols)
// 	inStr := prepString(INS, table, cols, "")

// 	stmt, err := db.Prepare(inStr)
// 	if err != nil {
// 		dbFail(err.Error())
// 	}

// 	faceCols := toInterfaceSlice(cols)
// 	_, err = stmt.Exec(faceCols...)

// 	return err
// }

func (db *DB) Select(table string, cols []string, condition string, vals []interface{}) (*sql.Rows, error) {
	prepStr := prepString(SEL, table, cols, condition)

	stmt, err := db.Prepare(prepStr)
	if err != nil {
		dbFail(err.Error())
	}

	rows, err := stmt.Query(vals...)
	if err != nil {
		dbFail(err.Error())
	}

	return rows, err
}

func (db *DB) Insert(table string, cols []string, vals []interface{}) (sql.Result, error) {

	return db.sqlGeneric(INS, table, cols, vals, "", nil)
}

func (db *DB) Update(table string, cols []string, condition string, vals []interface{}) (sql.Result, error) {

	return db.sqlGeneric(UPD, table, cols, vals, condition, nil)
}

func (db *DB) Delete(table string, conditions string, vals []interface{}) (sql.Result, error) {

	return db.sqlGeneric(DEL, table, nil, vals, conditions, nil)
}

func Cols(columns ...string) []string {
	return columns
}

func Vals(values ...interface{}) []interface{} {
	return values
}

func (db *DB) sqlGeneric(crud int, table string, cols []string,
	vals []interface{}, condition string, rows *sql.Rows) (sql.Result, error) {

	prepStr := prepString(crud, table, cols, condition)

	stmt, err := db.Prepare(prepStr)
	if err != nil {
		dbFail(err.Error())
	}

	res, err := stmt.Exec(vals...)
	if err != nil {
		dbFail(err.Error())
	}

	return res, err
}

// func (db *DB) sqlGeneric() func(int, string, []string, []interface{}, string, *sql.Rows) (sql.Result, error) {

// 	return func(crud int, table string, cols []string,
// 		vals []interface{}, condition string, rows *sql.Rows) (sql.Result, error) {

// 		prepStr := prepString(crud, table, cols, condition)

// 		stmt, err := db.Prepare(prepStr)
// 		if err != nil {
// 			dbFail(err.Error())
// 		}

// 		res, err := stmt.Exec(vals...)
// 		if err != nil {
// 			dbFail(err.Error())
// 		}

// 		return res, err
// 	}
// }

// int, string, []string, []string, string

// Helper Function: Returns an insert statment as a string.
// func prepString(crud string, table string, cols []string) string {

// 	return "INSERT INTO " + table +

// 		"(" +
// 		strings.Join(cols, ", ") +
// 		") " +

// 		"VALUES " +

// 		"(" +
// 		safeMarkers(len(cols)) +
// 		") "
// }

func prepString(crud int, table string, cols []string, condition string) string {
	retStr := make([]string, CRUD_MAX)
	copy(retStr, sqlCRUD[crud])

	colString := ""
	if crud != DEL {
		colString = strings.Join(cols, ", ")
	}

	switch crud {
	case INS:
		// ["INSERT INTO", "", "VALUES", ""]
		retStr[1] = table + " (" + colString + ")"
		retStr[3] = safeMarkers(len(cols))

	case SEL:
		// ["SELECT (", "", ") FROM", "", "WHERE", ""]
		retStr[1] = colString
		retStr[3] = table
		retStr[5] = condition

	case UPD:
		// ["UPDATE", "", "SET (", "", ")WHERE", ""]
		retStr[1] = table
		retStr[3] = colString
		retStr[5] = condition

	case DEL:
		// ["DELETE FROM", "", "WHERE", ""]
		retStr[1] = table
		retStr[3] = condition
	}

	return strings.Join(retStr, "")
}

func safeMarkers(num int) string {
	s := ""
	for i := 0; i < num-1; i++ {
		s += "?, "
	}
	s += "?"

	return s
}

// func toInterfaceSlice(sli []string) []interface{} {
// 	var newSli []interface{}
// 	for _, s := range sli {
// 		newSli = append(newSli, s)
// 	}

// 	return newSli
// }

// func dbfail() func(...interface{}) {
// 	return log.Fatalln
// }
