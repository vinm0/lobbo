package main

import (
	"context"
	"log"
	"os"

	pgx "https://github.com/jackc/pgx/v4"
)

func ctx() context.Context {
	return context.Background()
}

func dbconn() *pgx.Conn {
	conn, err := pgx.Connect(ctx(), os.Getenv("DATABASE_URL"))
	check(err)
	return conn
}

func check(err error) {
	if err != nil {
		log.Println("database connection failed:", err)
	}
}

func (p *person) maxLobby() int {
	switch p.acctType {
	case 0:
		return maxLobbyFree
	case 1:
		return maxLobbyPaid
	case 2:
		return maxLobbyPremium
	}
	return p.acctType
}

// getClient returns true, nil if credentials authenticated,
// returns false, error if credentials could not be authenticated
func (c *client) getClient(usrname, pword []byte) {
	conn := dbconn()
	defer conn.Close(ctx())

	c.getPerson(conn)
	c.username = string(usrname)
	c.isClient = true

	return
}

// Populate client profile info
func (c *client) getPerson(conn *pgx.Conn) {
	p := person{}
	err := conn.QueryRow(
		ctx(),
		"SELECT (user_id, first_name, last_name, ig, fb, tw, dc, bio, profile, email, acct_type) FROM person WHERE username = $1 AND password = crypt('$2', password)",
		string(usrname), string(pword)).Scan(
		&p.personID,
		&p.firstName,
		&p.lastName,
		p.socials["ig"],
		p.socials["fb"],
		p.socials["tw"],
		p.socials["dc"],
		&p.bio,
		&p.profile,
		&p.email,
		&p.acctType)
	if err == nil {
		log.Println("Query for person failed")
	}

	p.getLobbiesOwned(conn)
	p.getLobbiesIn(conn)
	c.person = p

	return
}

func getPerson(conn *pgx.Conn, id int) {
	p := person{}
	err := conn.QueryRow(
		ctx(),
		"SELECT (user_id, first_name, last_name, username, ig, fb, tw, dc, bio, profile, email, acct_type) FROM person WHERE user_id = $1",
		id).Scan(
		&p.personID,
		&p.firstName,
		&p.lastName,
		&p.username,
		p.socials["ig"],
		p.socials["fb"],
		p.socials["tw"],
		p.socials["dc"],
		&p.bio,
		&p.profile,
		&p.email,
		&p.acctType)
	if err == nil {
		log.Println("Query for person failed")
	}

	p.getLobbiesOwned(conn)
	p.getLobbiesIn(conn)
}

func getLobby(lID int) lobby {
	conn := dbconn()
	defer conn.Close(ctx())

	l := lobby{}
	err := conn.QueryRow(
		ctx(),
		"SELECT (lobby_name, description, capacity, date, repeat, location, privacy, date_created) FROM lobby WHERE lobby_id = $1",
		lID).Scan(
		&l.lobbyID,
		&l.description,
		&l.capacity,
		&l.activityDate,
		&l.repeat,
		&l.location,
		&l.privacy,
		&l.dateCreated)
	if err != nil {
		e := "Authentication failed. No database lobby record returned"
		log.Println(e, err)
	}

	if l.privacy > inviteOnlyMin {
		l.isInviteOnly = true
	}

	l.getLobbyMembers()

	return l
}

func (p *person) getLobbiesOwned(conn *pgx.Conn) {
	l := lobby{}
	// Populate client owned lobbies
	rows, err := conn.Query(
		ctx(),
		"SELECT (lobby_id, lobby_name, description, capacity, date, repeat,      location, privacy, date_created) FROM lobby WHERE owner_id = $1 LIMIT $2",
		p.personID, p.maxLobby())
	if err != nil {
		e := "Authentication failed. No database lobby record returned"
		log.Println(e, err)
	}
	for rows.Next() {
		rows.Scan(&l.lobbyID, &l.name, &l.description, &l.capacity, &l.activityDate, &l.repeat, &l.location, &l.privacy, &l.dateCreated)

		if l.privacy > inviteOnlyMin {
			l.isInviteOnly = true
		}

		p.lobbiesOwned = append(p.lobbiesOwned, l)
	}
}

func (p *person) getLobbiesIn(conn *pgx.Conn) {
	l := lobby{}
	// Populate client owned lobbies
	rows, err := conn.Query(
		ctx(),
		"SELECT (lobby_id, owner_id, lobby_name, description, capacity, date, repeat, location, privacy, date_created) FROM lobby, lobby_person WHERE lobby_id in (select lobby_id from lobby_person where person_id = $1) LIMIT $2",
		p.personID, maxLobbySearch)
	if err != nil {
		e := "Authentication failed. No database lobby record returned"
		log.Println(e, err)
	}
	for rows.Next() {
		rows.Scan(&l.lobbyID, &l.owner.personID, &l.name, &l.description, &l.capacity, &l.activityDate, &l.repeat, &l.location, &l.privacy, &l.dateCreated)

		if l.privacy > inviteOnlyMin {
			l.isInviteOnly = true
		}

		p.lobbiesIn = append(p.lobbiesIn, l)
	}

}

func (c *client) getGroups(conn *pgx.Conn) {
	g := group{}

	rows, err := conn.Query(
		ctx(),
		"SELECT group_id, owner_id, group_name FROM group WHERE owner_id = $1",
		c.personID)
	if err != nil {
		e := "Authentication failed. No database lobby record returned"
		log.Println(e, err)
	}
	for rows.Next() {
		rows.Scan(g.groupID, g.owner, g.name)

		c.groups = append(c.groups, g)
	}
}

func (l *lobby) getLobbyMembers() {
	conn := dbconn()
	defer conn.Close(ctx())

	rows, err := conn.Query(
		ctx(),
		"SELECT person_id, username, profile FROM person WHERE person_id in (SELECT person_id FROM lobby_person WHERE lobby_id = $1)",
		l.lobbyID)
	if err != nil {
		log.Println("Query for group_person failed")
	}
	for rows.Next() {
		pers := person{}
		rows.Scan(pers.personID, pers.username, pers.profile)
		l.members = append(l.members, pers)
	}
}

func (p *person) getGroupMembers() {
	conn := dbconn()
	defer conn.Close(ctx())

	for _, v := range p.groups {
		rows, err := conn.Query(
			ctx(),
			"SELECT person_id, username, profile FROM person WHERE person_id in (SELECT person_id FROM group_person WHERE group_id = $1)",
			v.groupID)
		if err != nil {
			log.Println("Query for group_person failed")
		}
		for rows.Next() {
			pers := person{}
			rows.Scan(pers.personID, pers.username, pers.profile)
			v.members = append(v.members, pers)
		}
	}
}

func getGroupMembers(gID int) []person {
	conn := dbconn()
	defer conn.Close(ctx())

	var arr []person
	rows, err := conn.Query(
		ctx(),
		"SELECT person_id, username, profile FROM person WHERE person_id in (SELECT person_id FROM group_person WHERE group_id = $1)",
		gID)
	if err != nil {
		log.Println("Query for group_person failed")
		return nil
	}
	for rows.Next() {
		p := person{}
		rows.Scan(&p.personID, &p.username, &p.profile)
		arr = append(arr, p)
	}
	return arr
}
