package main

import "errors"

// A Group represents a collection of Leaders.
// Groups should only be accessible to the Group's owner.
type Group struct {
	GroupID int       // Unique id of the Group
	OwnerID int       // ID of the Group's owner
	Name    string    // Name of the Group
	Members []*Leader // Members of the Group added by the owner.
}

// Deletes the Group from the database.
// Returns an error if the leader referenced is not the Group's owner
func (g *Group) Delete(leaderID int) error {
	if leaderID != g.OwnerID {
		return errors.New("the leaderID provided does not match the Group's OwnerID")
	}

	DeleteGroupDB(g.GroupID)
	return nil
}
