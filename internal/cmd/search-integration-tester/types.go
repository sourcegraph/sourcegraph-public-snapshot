package main

type GQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type GQLSearchVariable struct {
	SearchQuery string `json:"query"`
}

type GQLResult interface{}
