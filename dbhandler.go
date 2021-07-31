package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_DRIVER = "sqlite3"
	DB_PATH   = "./database/sample-lobbo.db"

	// DB CRUD
	INS      = 0
	SEL      = 1
	UPD      = 2
	DEL      = 3
	CRUD_MAX = 6 // Longest length of sqlCRUD slice

	// Error Messages
	CONN_FAIL = "Unable to connect to Database."
)

var (
	dbFail func(...interface{}) = log.Println

	// SQL command templates
	sqlCRUD = [][]string{
		{"INSERT INTO ", "", " VALUES ", ""},
		{"SELECT ", "", " FROM ", "", " WHERE ", ""},
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

func (db *DB) Select(table string, cols []string, condition string, vals ...interface{}) (*sql.Rows, error) {
	prepStr := prepString(SEL, table, cols, condition)

	fmt.Println("String prepped:", prepStr)
	stmt, err := db.Prepare(prepStr)
	if err != nil {
		dbFail(err.Error())
	}
	defer stmt.Close()

	fmt.Println("Statement created")
	rows, err := stmt.Query(vals...)
	if err != nil {
		dbFail(err.Error())
	}

	fmt.Println("returning rows")
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

func Auth(usr string, pwd string) (client *Leader, err error) {
	db, err := ConnectDB()
	if err != nil {
		dbFail("Cannot connect to database", err)
		return nil, err
	}

	defer db.Close()

	prep := "SELECT leader_id, fname, lname FROM leaders WHERE usrname = ? AND pwd = ?"
	stmt, err := db.Prepare(prep)
	if err != nil {
		dbFail(err.Error())
		return nil, err
	}

	var id int
	var fname string
	var lname string

	if err = stmt.QueryRow(usr, pwd).
		Scan(&id, &fname, &lname); err != nil {
		dbFail(err.Error())
		return nil, err
	}

	return &Leader{
		LeaderID:  id,
		Username:  usr,
		Firstname: fname,
		Lastname:  lname}, nil
}

func OwnedLobbiesDB(ownerID int, limit string) []*Lobby {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"lobby_id", "title", "lobby_desc", "meet_time"}
	condition := "owner_id = ?" + limit

	rows, err := db.Select("lobbies", cols, condition, ownerID)
	Check(err, "Unable to query lobbies owned for ownerID", ownerID)

	return loadOwnedLobbies(rows)
}

func inLobbiesDB(memberID int, limit string) []*Lobby {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	ownerName := "fname||' '||lname"
	cols := []string{"l.lobby_id", ownerName, "title", "lobby_desc", "meet_time"}
	condition := "member_id = ?" + limit
	table := "lobbies l JOIN leaders ON leader_id = owner_id JOIN lobby_members lm ON l.lobby_id = lm.lobby_id"

	rows, err := db.Select(table, cols, condition, memberID)
	Check(err, "Unable to query lobbies owned for ownerID", memberID)

	return loadInLobbies(rows)
}

func loadOwnedLobbies(rows *sql.Rows) []*Lobby {
	lobbies := []*Lobby{}
	for rows.Next() {
		l := Lobby{}
		var meetTime string

		rows.Scan(&l.LobbyID, &l.Title, &l.Description, &meetTime)

		if meetTime != "" {
			t, err := time.Parse(time.RFC822, meetTime)
			Check(err, "Unable to parse meeting time:", meetTime)

			l.MeetTime = t
		}

		lobbies = append(lobbies, &l)
	}

	return lobbies
}

func loadInLobbies(rows *sql.Rows) []*Lobby {
	lobbies := []*Lobby{}
	for rows.Next() {
		l := Lobby{}
		var meetTime string

		rows.Scan(&l.LobbyID, &l.OwnerName, &l.Title, &l.Description, &meetTime)

		if meetTime != "" {
			t, err := time.Parse(time.RFC822, meetTime)
			Check(err, "Unable to parse meeting time:", meetTime)

			l.MeetTime = t
		}

		lobbies = append(lobbies, &l)
	}

	return lobbies
}
