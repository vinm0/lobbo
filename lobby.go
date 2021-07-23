package main

type Lobby struct {
	OwnerID     int
	Members     []*Leader
	Description string
	Capacity    int
	Privacy     int
	Visibility  int
}
