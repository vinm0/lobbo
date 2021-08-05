package main

import "strconv"

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

func (ldr *Leader) ownsGroup(groupID int) bool {
	return ldr.LeaderID == GroupOwnerDB(groupID)
}

func (ldr *Leader) Groups() []Group {
	return GroupsDB(ldr.LeaderID)
}

func (ldr *Leader) Colleagues(limit int) []*Leader {
	return ColleaguesDB(ldr.LeaderID, " Limit "+strconv.Itoa(limit))
}

func (ldr *Leader) ColleaguesAll() []*Leader {
	return ColleaguesDB(ldr.LeaderID, "")
}
