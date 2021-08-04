package main

type Leader struct {
	LeaderID     int
	Username     string
	Firstname    string
	Lastname     string
	OwnedLobbies []*Lobby
}

func (ldr *Leader) isOwner(lby *Lobby) bool {
	return ldr.LeaderID == lby.OwnerID
}
