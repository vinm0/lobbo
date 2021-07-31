package main

import "time"

type Lobby struct {
	LobbyID     int
	OwnerID     int
	OwnerName   string
	Title       string
	Members     []*Leader
	Description string
	Capacity    int
	Privacy     int
	Visibility  int
	MeetTime    time.Time
}
