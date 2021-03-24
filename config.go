package main

type Config struct {
	Groups []QueryGroupConfig
}

type QueryGroupConfig struct {
	Name    string
	Queries []QueryConfig
}

type QueryConfig struct {
	Query string
	Name  string
}
