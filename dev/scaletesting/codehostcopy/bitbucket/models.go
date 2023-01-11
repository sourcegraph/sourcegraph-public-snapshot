package bitbucket

import "encoding/json"

type PagedResp struct {
	Start         int               `json:"start"`
	Size          int               `json:"size"`
	Limit         int               `json:"Limit"`
	IsLastPage    bool              `json:"isLastPage,omitempty"`
	NextPageStart int               `json:"nextPageStart,omitempty"`
	Values        []json.RawMessage `json:"values"`
}

type Href struct {
	Url  string `json:"href"`
	Name string `json:"name,omitempty"`
}

type Repo struct {
	Id      int               `json:"id"`
	Links   map[string][]Href `json:"links"`
	Name    string            `json:"name"`
	Project Project           `json:"project"`
	ScmId   string            `json:"scmId"`
	Slug    string            `json:"slug"`
	State   string            `json:"state"`
}

type Project struct {
	Key    string `json:"key"`
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"bool"`
	Type   string `json:"type"`
}

type User struct{}
type Group struct{}

type State struct {
	Users    []User    `json:"users"`
	Groups   []Group   `json:"groups"`
	Repos    []Repo    `json:"repos"`
	Projects []Project `json:"projects"`
}
