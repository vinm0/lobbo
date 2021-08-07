package main

type Group struct {
	GroupID int
	OwnerID int
	Name    string
	Members []*Leader
}

func (g *Group) Delete() {
	DeleteGroupDB(g.GroupID)
}
