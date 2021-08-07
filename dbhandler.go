package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
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

	TIME_FORMAT = "2006-01-02 15:04"
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
	prepStr := prepString(SEL, table, cols, condition, len(vals))

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

func (db *DB) Insert(table string, cols []string, vals ...interface{}) (sql.Result, error) {

	return db.sqlGeneric(INS, table, cols, vals, "", nil)
}

func (db *DB) InsertMany(table string, cols []string, vals ...interface{}) (sql.Result, error) {

	return db.sqlGeneric(INS, table, cols, vals, "", nil)
}

func (db *DB) Update(table string, cols []string, condition string, vals ...interface{}) (sql.Result, error) {
	cols = sanatizeCols(cols)
	return db.sqlGeneric(UPD, table, cols, vals, condition, nil)
}

func (db *DB) Delete(table string, conditions string, vals ...interface{}) (sql.Result, error) {

	return db.sqlGeneric(DEL, table, nil, vals, conditions, nil)
}

func Cols(columns ...string) []string {
	return columns
}

func Vals(values ...interface{}) []interface{} {
	return values
}

func sanatizeCols(cols []string) []string {
	for i, v := range cols {
		cols[i] = v + " = ?"
	}
	return cols
}

func (db *DB) sqlGeneric(crud int, table string, cols []string,
	vals []interface{}, condition string, rows *sql.Rows) (sql.Result, error) {

	prepStr := prepString(crud, table, cols, condition, len(vals))
	fmt.Println("String prepped for insert")
	fmt.Println(prepStr)

	stmt, err := db.Prepare(prepStr)
	if err != nil {
		dbFail(err.Error())
	}

	fmt.Println("Executing sql statement")
	res, err := stmt.Exec(vals...)
	if err != nil {
		dbFail(err.Error())
	}
	fmt.Println("Statement Executed")

	return res, err
}

func prepString(crud int, table string, cols []string, condition string, valsLen int) string {
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
		retStr[3] = safeMarkers(len(cols), valsLen)

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

func safeMarkers(colsLen int, valsLen int) string {
	b := strings.Builder{}

	for row := 0; row < (valsLen / colsLen); row++ {
		b.WriteString("(")

		for i := 1; i < colsLen; i++ {
			b.WriteString("?,")
		}
		b.WriteString("?),")
	}

	return strings.TrimSuffix(b.String(), ",")
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

	cols := []string{"lobby_id", "title", "summary", "meet_time", "meet_loc"}
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
	cols := []string{"l.lobby_id", ownerName, "title", "summary", "meet_time", "meet_loc"}
	condition := "member_id = ?" + limit
	table := "lobbies l JOIN leaders ON leader_id = owner_id JOIN lobby_members lm ON l.lobby_id = lm.lobby_id"

	rows, err := db.Select(table, cols, condition, memberID)
	Check(err, "Unable to query lobbies owned for ownerID", memberID)

	return loadInLobbies(rows)
}

func loadOwnedLobbies(rows *sql.Rows) []*Lobby {
	defer rows.Close()

	lobbies := []*Lobby{}
	for rows.Next() {
		l := Lobby{}
		var meetTime string

		rows.Scan(&l.LobbyID, &l.Title, &l.Description, &meetTime, &l.Location)

		if meetTime != " " {
			t, err := time.Parse(TIME_FORMAT, meetTime)
			Check(err, "Unable to parse meeting time:", meetTime)

			l.MeetTime = t
		}
		fmt.Println(l)
		lobbies = append(lobbies, &l)
	}
	return lobbies
}

func loadInLobbies(rows *sql.Rows) []*Lobby {
	defer rows.Close()

	lobbies := []*Lobby{}
	for rows.Next() {
		l := Lobby{}
		var meetTime string

		rows.Scan(&l.LobbyID, &l.OwnerName, &l.Title,
			&l.Description, &meetTime, &l.Location)

		if meetTime != " " {
			t, err := time.Parse(TIME_FORMAT, meetTime)
			Check(err, "Unable to parse meeting time:", meetTime)

			l.MeetTime = t
		}

		lobbies = append(lobbies, &l)
	}

	return lobbies
}

func ColleaguesDB(ownerID int, limit string) []*Leader {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"colleague_id", "fname", "lname", "usrname"}
	condition := "owner_id = ? ORDER BY fname, lname " + limit
	table := "colleagues JOIN leaders ON colleague_id = leader_id"
	rows, err := db.Select(table, cols, condition, ownerID)
	Check(err, "Unable to query lobbies owned for ownerID", ownerID)

	return loadLeaders(rows)
}

func loadLeaders(rows *sql.Rows) []*Leader {
	defer rows.Close()

	leaders := []*Leader{}
	for rows.Next() {
		l := Leader{}
		rows.Scan(&l.LeaderID, &l.Firstname, &l.Lastname, &l.Username)

		leaders = append(leaders, &l)
	}

	return leaders
}

func DeleteColleagueDB(ownerID int, colleagueID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	condition := "owner_id = ? AND colleague_id = ?"
	_, err = db.Delete("colleagues", condition, ownerID, colleagueID)
	Check(err, "Unable to query lobbies owned for ownerID", ownerID)
}

func LobbyDB(id int) *Lobby {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"*"}
	rows, err := db.Select("lobbies", cols, "lobby_id = ?", id)
	Check(err, "Unable to query lobby for id", id)

	return loadLobby(rows)
}

func loadLobby(rows *sql.Rows) *Lobby {
	defer rows.Close()
	l := Lobby{}
	var meetTime string
	if rows.Next() {
		rows.Scan(&l.LobbyID, &l.OwnerID, &l.Title, &l.Description,
			&meetTime, &l.Location, &l.Link, &l.Capacity,
			&l.Visibility, &l.InviteOnly)
	}

	if meetTime != " " {
		t, err := time.Parse(TIME_FORMAT, meetTime)
		Check(err, "Unable to parse meeting time:", meetTime)

		l.MeetTime = t
	}

	return &l
}

func LeaderDB(leaderID int) *Leader {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"leader_id", "fname", "lname", "usrname"}
	condition := "leader_id = ?"

	rows, err := db.Select("leaders", cols, condition, leaderID)
	Check(err, "Unable to query leaders for leaderID", leaderID)

	return loadLeader(rows)
}

func loadLeader(row *sql.Rows) *Leader {
	defer row.Close()
	l := Leader{}
	if row.Next() {
		row.Scan(&l.LeaderID, &l.Firstname, &l.Lastname, &l.Username)
	}

	return &l
}

func MembersDB(lobbyID int) []*Leader {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"member_id", "fname", "lname", "usrname"}
	condition := "lobby_id = ?"
	table := "lobby_members JOIN leaders ON member_id = leader_id"

	rows, err := db.Select(table, cols, condition, lobbyID)
	Check(err, "Unable to query members for lobbyID", lobbyID)

	return loadLeaders(rows)
}

func JoinLobbyDB(lobbyID int, leaderID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{"lobby_id", "member_id"}

	_, err = db.Insert("lobby_members", cols, lobbyID, leaderID)
	Check(err, "Unable to add leader to lobby ", lobbyID)

}

func CreateLobbyDB(form url.Values) (newLobbyID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{
		"owner_id", "title", "summary", "meet_time",
		"meet_loc", "loc_link", "capacity", "visibility",
	}

	vals := formVals(cols, form)

	res, err := db.Insert("lobbies", cols, vals...)
	Check(err, "Unable to create lobby ", form["lobby_id"])

	id, _ := res.LastInsertId()
	newLobbyID = int(id)

	return newLobbyID
}

func UpdateLobbyDB(form url.Values, lobby_id int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := []string{
		"title", "summary", "meet_time", "meet_loc",
		"loc_link", "capacity", "visibility",
	}

	vals := formVals(cols, form)
	vals = append(vals, lobby_id)
	fmt.Println("vals: ", vals)
	condition := "lobby_id = ?"

	_, err = db.Update("lobbies", cols, condition, vals...)
	Check(err, "Unable to update lobby ", lobby_id)
}

func GroupNameDB(groupID int) string {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("groupname")

	rows, err := db.Select("groups", cols, "group_id = ?", groupID)
	Check(err, "Unable to select groupname for group ", groupID)

	return loadGroupName(rows)
}

func loadGroupName(rows *sql.Rows) string {
	defer rows.Close()

	var name string
	if rows.Next() {
		rows.Scan(&name)
	}

	return name
}

func GroupsDB(owner_id int) []Group {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	s := `SELECT group_id, groupname, owner_id, NULL, NULL, NULL, NULL
	FROM groups
	WHERE owner_id = ?
	
	UNION
	
	SELECT g.group_id, groupname, owner_id, leader_id, fname, lname, usrname 
	FROM groups g 
			JOIN group_members gm 
				ON (g.group_id = gm.group_id) 
			JOIN leaders 
				ON (member_id = leader_id) 
	WHERE owner_id = ? 
	ORDER BY g.group_id;`

	stmt, err := db.Prepare(s)
	Check(err, "invalid statement", s)

	rows, err := stmt.Query(owner_id, owner_id)
	Check(err, "unable to query groups for user ", owner_id)

	// cols := []string{
	// 	"g.group_id", "groupname", "owner_id",
	// 	"leader_id", "fname", "lname", "usrname",
	// }

	// table := "groups g JOIN group_members gm ON (g.group_id = gm.group_id) " +
	// 	"JOIN leaders ON (member_id = leader_id)"

	// condition := "owner_id = ? ORDER BY g.group_id"

	// rows, err := db.Select(table, cols, condition, owner_id)
	// Check(err, "Unable to select groups for ", owner_id)

	return loadGroups(rows)
}

func loadGroups(rows *sql.Rows) []Group {
	defer rows.Close()

	groups := []Group{}

	var g Group
	currGID := 0

	for rows.Next() {
		var gID, ownID int
		var gName string
		ldr := Leader{}

		rows.Scan(&gID, &gName, &ownID,
			&ldr.LeaderID, &ldr.Firstname, &ldr.Lastname, &ldr.Username)

		// new group scanned
		if gID != currGID {
			if currGID != 0 {
				groups = append(groups, g)
			}

			g = Group{GroupID: gID, Name: gName, OwnerID: ownID}
		}

		g.Members = append(g.Members, &ldr)
		currGID = g.GroupID
	}

	groups = append(groups, g)
	return groups
}

func formVals(cols []string, form url.Values) (vals []interface{}) {
	vals = []interface{}{}
	for _, v := range cols {
		if v == "visibility" {
			num, _ := strconv.Atoi(form.Get(v))
			vals = append(vals, num)
			continue
		}
		vals = append(vals, form.Get(v))
	}
	return vals
}

func GroupOwnerDB(groupID int) (ownerID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("owner_id")

	row, err := db.Select("groups", cols, "group_id = ?", groupID)
	Check(err, "Unable to select owner from group ", groupID)

	return loadLeaderID(row)
}

func loadLeaderID(row *sql.Rows) int {
	defer row.Close()

	var id int
	if row.Next() {
		row.Scan(&id)
	}

	return id
}

func DeleteGroupMemberDB(groupID int, memberID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	condition := "group_id = ? AND member_id = ?"
	_, err = db.Delete("group_members", condition, groupID, memberID)
	Check(err, "Unable to delete member ", memberID, " from group ", groupID)
}

func CreateGroupDB(form url.Values) (newID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("owner_id", "groupname")
	id, _ := strconv.Atoi(form.Get("owner_id"))
	name := form.Get("groupname")

	res, err := db.Insert("groups", cols, id, name)
	Check(err, "Unable to create group ", name)

	id64, _ := res.LastInsertId()

	return int(id64)
}

func UpdateGroupDB(form url.Values, groupID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("groupname")

	vals := formVals(cols, form)
	vals = append(vals, groupID)

	condition := "group_id = ?"

	_, err = db.Update("groups", cols, condition, vals...)
	Check(err, "Unable to update group ", form.Get("groupname"))
}

func AddGroupMembersDB(form url.Values, groupID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("group_id", "member_id")
	vals := fromIDString(groupID, form["colleagues"])

	_, err = db.InsertMany("group_members", cols, vals...)
	Check(err, "Unable to add members to group_members")
}

func fromIDString(groupID int, ids []string) []interface{} {
	s := make([]interface{}, len(ids)*2)
	for _, v := range ids {
		s = append(s, groupID)
		num, _ := strconv.Atoi(v)
		s = append(s, num)
	}

	return s
}

func DeleteLobbyMemberDB(lobbyID int, memberID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	fmt.Println("lobbyid: ", lobbyID)
	fmt.Println("memberid: ", memberID)
	condition := "lobby_id = ? AND member_id = ?"
	_, err = db.Delete("lobby_members", condition, lobbyID, memberID)
	Check(err, "Unable to delete member ", memberID, " from lobby ", lobbyID)
}

func AddColleagueDB(ownerID int, colleagueID int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("owner_id", "colleague_id")

	_, err = db.Insert("colleagues", cols, ownerID, colleagueID)
	Check(err, "Unable to add colleague", colleagueID, " for ", ownerID)
}

func ColleagueDB(leaderID int, colleagueID int) *Leader {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	cols := Cols("leader_id", "fname", "lname", "usrname")
	table := "colleagues JOIN leaders ON (colleague_id = leader_id)"
	condition := "owner_id = ? AND colleague_id = ?"
	row, err := db.Select(table, cols, condition, leaderID, colleagueID)
	Check(err, "Unable to check for colleague ", colleagueID)

	return loadLeader(row)
}

func DeleteLobbyDB(id int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	_, err = db.Delete("lobbies", "lobby_id = ?", id)
	Check(err, "Unable to delete lobby ", id)
}

func DeleteGroupDB(id int) {
	db, err := ConnectDB()
	Check(err, CONN_FAIL)
	defer db.Close()

	_, err = db.Delete("groups", "group_id = ?", id)
	Check(err, "Unable to delete group ", id)
}

// TODO: validate join permissions based on invite code.
// func JoinAllowed(lobbyID int, leaderID int, privacy int) bool {
// 	db, err := ConnectDB()
// 	Check(err, CONN_FAIL)
// 	defer db.Close()

// 	cols := []string{"member_id", "fname", "lname", "usrname"}
// 	condition := "lobby_id = ?"
// 	table := "lobby_members JOIN leaders ON member_id = leader_id"

// 	rows, err := db.Select(table, cols, condition, lobbyID)
// 	Check(err, "Unable to query members for lobbyID", lobbyID)

// }
