package testdata

import "embed"

//go:embed **/*
var Content embed.FS
