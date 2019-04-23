package main

const orgFragment = `
fragment OrgFields on Org {
    id
    name
    displayName
    members {
        nodes {
			id
			username
		}
    }
}
`

type Org struct {
	ID          string
	Name        string
	DisplayName string
	Members     struct {
		Nodes []User
	}
}
