package main

const userFragment = `
fragment UserFields on User {
    id
    username
    displayName
    siteAdmin
    organizations {
		nodes {
        	id
        	name
        	displayName
		}
    }
    emails {
        email
        verified
    }
    url
}
`

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
