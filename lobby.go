package main

import "time"

type Lobby struct {
	LobbyID     int
	OwnerID     int
	OwnerName   string
	Title       string
	Members     []*Leader
	Description string
	Location    string
	Link        string
	Capacity    int
	Privacy     int
	Visibility  int
	InviteOnly  int
	MeetTime    time.Time
}
