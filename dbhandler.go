package main

import (
	"database/sql"
	"log"
	"strings"
)

const (
	DB_DRIVER = "sqlite3"
	DB_PATH   = "./sample-lobbo.db"

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

func (db *DB) Insert(table string, cols []string, vals []interface{}) (sql.Result, error) {
	insFunc := db.sqlGeneric()

	return insFunc(INS, table, cols, vals, "")
}

func (db *DB) Select(table string, cols []string, condition string, vals []interface{}) (sql.Result, error) {
	selFunc := db.sqlGeneric()

	return selFunc(SEL, table, cols, vals, condition)
}

func (db *DB) Update(table string, cols []string, condition string, vals []interface{}) (sql.Result, error) {
	updFunc := db.sqlGeneric()

	return updFunc(UPD, table, cols, vals, condition)
}

func (db *DB) Delete(table string, conditions string, vals []interface{}) (sql.Result, error) {
	delFunc := db.sqlGeneric()

	return delFunc(DEL, table, nil, vals, conditions)
}

func (db *DB) sqlGeneric() func(int, string, []string, []interface{}, string) (sql.Result, error) {

	return func(crud int, table string, cols []string,
		vals []interface{}, condition string) (sql.Result, error) {

		prepStr := prepString(crud, table, cols, condition)

		stmt, err := db.Prepare(prepStr)
		if err != nil {
			dbFail(err.Error())
		}

		return stmt.Exec(vals...)
	}
}

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
