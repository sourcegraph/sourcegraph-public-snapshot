package mmark

import (
	"bytes"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

type author struct {
	Initials     string
	Surname      string
	Fullname     string
	Organization string
	Role         string
	Ascii        string
	Address      address
}

type address struct {
	Phone  string
	Email  string
	Uri    string
	Postal addressPostal
}

type addressPostal struct {
	Street     string
	City       string
	Code       string
	Country    string
	PostalLine []string
}

type title struct {
	Title  string
	Abbrev string

	DocName        string
	Ipr            string
	Category       string
	Obsoletes      []string
	Updates        []string
	SubmissionType string

	Date      time.Time
	Area      string
	Workgroup string
	Keyword   []string
	Author    []author
}

func (p *parser) titleBlockTOML(out *bytes.Buffer, data []byte) title {
	data = bytes.TrimPrefix(data, []byte("%"))
	data = bytes.Replace(data, []byte("\n%"), []byte("\n"), -1)
	var block title
	if _, err := toml.Decode(string(data), &block); err != nil {
		log.Printf("mmark: TOML titleblock: %s", err.Error())
		return block // never an error when encoding markdown
	}
	return block
}
