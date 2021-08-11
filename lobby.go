package main

import "time"

// A Lobby contains the details of a lobby created and maintained by a leader
type Lobby struct {
	LobbyID     int
	OwnerID     int
	OwnerName   string // Full name of the Lobby's owner
	Title       string
	Members     []*Leader
	Description string
	Location    string
	Link        string // The web link assocaited with the location (GPS, website)
	Capacity    int
	Privacy     int
	Visibility  int
	InviteOnly  int
	MeetTime    time.Time
}

// Returns the owner of the assocaited Lobby
func (l *Lobby) Owner() *Leader {
	return LeaderDB(l.OwnerID)
}

// Deletes the Lobby from the database
func (l *Lobby) Delete() {
	DeleteLobbyDB(l.LobbyID)
}
