package main

type User struct {
	ID            string
	Username      string
	DisplayName   string
	SiteAdmin     bool
	Organizations struct {
		Nodes []Org
	}
	Emails []UserEmail
	URL    string
}

type UserEmail struct {
	Email    string
	Verified bool
}

type Org struct {
	ID          string
	Name        string
	DisplayName string
	Members     struct {
		Nodes []User
	}
}
