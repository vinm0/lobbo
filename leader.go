package main

import "strconv"

type Leader struct {
	LeaderID     int
	Username     string
	Firstname    string
	Lastname     string
	OwnedLobbies []*Lobby
}

func (ldr *Leader) OwnsLobby(lbyID string) bool {
	id, _ := strconv.Atoi(lbyID)
	return ldr.LeaderID == LobbyDB(id).OwnerID
}

func (ldr *Leader) ownsGroup(groupID int) bool {
	return ldr.LeaderID == GroupOwnerDB(groupID)
}

func (ldr *Leader) Groups() []Group {
	return GroupsDB(ldr.LeaderID)
}

func (ldr *Leader) GroupName(groupID string) string {
	id, _ := strconv.Atoi(groupID)
	return GroupNameDB(id)
}

func (ldr *Leader) Colleagues(limit int) []*Leader {
	return ColleaguesDB(ldr.LeaderID, " Limit "+strconv.Itoa(limit))
}

func (ldr *Leader) ColleaguesAll() []*Leader {
	return ColleaguesDB(ldr.LeaderID, "")
}
