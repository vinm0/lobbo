package main

import (
	"strconv"
)

// A Leader represents a Lobbo user.
type Leader struct {
	LeaderID  int
	Username  string
	Firstname string
	Lastname  string

	// Password will only be set for new accounts and sign-in.
	// Otherwise, Password will be empty
	Password string
}

// Creates a new account for the associated Leader.
func (ldr *Leader) CreateAccount() {
	AddLeaderDB(ldr)
}

// Checks whether the leader owns the lobby based on the Lobby's ID.
func (ldr *Leader) OwnsLobby(id string) bool {
	lbyID, _ := strconv.Atoi(id)
	return ldr.LeaderID == LobbyDB(lbyID).OwnerID
}

// Checks whether the leader owns the Group based on the Group's ID
func (ldr *Leader) ownsGroup(id string) bool {
	gID, _ := strconv.Atoi(id)
	return ldr.LeaderID == GroupOwnerDB(gID)
}

// Accepts the id of another Leader.
// Returns true if the receiver has added the other leader as a colleague.
// Returns false, otherwise. Also returns false if the receiver is the
// same Leader passed in the parameter.
func (ldr *Leader) IsColleague(id int) bool {
	if ldr.LeaderID == id {
		return false
	}

	l := ColleagueDB(ldr.LeaderID, id)
	return id == l.LeaderID
}

// Returns a slice of all Groups owned by the Leader.
func (ldr *Leader) Groups() []Group {
	return GroupsDB(ldr.LeaderID)
}

// Returns the name of a Group
func (ldr *Leader) GroupName(groupID string) string {
	id, _ := strconv.Atoi(groupID)
	return GroupNameDB(id)
}

// Adds a Leader to the receiver's list of Colleagues
func (ldr *Leader) AddColleague(colleagueID string) {
	id, _ := strconv.Atoi(colleagueID)
	AddColleagueDB(ldr.LeaderID, id)
}

// Returns a slice of Leaders representing the receiver's Colleagues.
// Use ColleaguesAll() method to retrieve all Colleagues.
func (ldr *Leader) Colleagues(limit int) []*Leader {
	return ColleaguesDB(ldr.LeaderID, " Limit "+strconv.Itoa(limit))
}

// Returns a slice of Leaders representing all the receiver's Colleagues
func (ldr *Leader) ColleaguesAll() []*Leader {
	return ColleaguesDB(ldr.LeaderID, "")
}
