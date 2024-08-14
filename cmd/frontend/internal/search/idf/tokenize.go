package idf

import (
	"regexp"
	"strings"
)

var (
	camelStartRe = regexp.MustCompile(`^[A-Za-z][^A-Z]+`)
	capStartRe   = regexp.MustCompile(`^[A-Z][A-Z0-9]*`)
)

func tokenizeCamelCase(s string) []string {
	remainder := s
	var toks []string
	for len(remainder) > 0 {
		if found := camelStartRe.FindString(remainder); found != "" {
			toks = append(toks, found)
			remainder = remainder[len(found):]
			continue
		}
		if found := capStartRe.FindString(remainder); found != "" {
			if len(found) == 1 || len(found) == len(remainder) {
				toks = append(toks, found)
				remainder = remainder[len(found):]
			} else {
				toks = append(toks, found[:len(found)-1])
				remainder = remainder[len(found)-1:]
			}
			continue
		}
		remainder = remainder[1:]
	}
	return toks
}

func tokenizeSnakeCase(s string) []string {
	return strings.Split(s, "_")
}

var (
	sepRe = regexp.MustCompile(`([[:punct:]]|\s)+`)
)

func TokenizeWord(w string) []string {
	var toks []string
	for _, part := range tokenizeSnakeCase(w) {
		toks = append(toks, tokenizeCamelCase(part)...)
	}
	return toks
}

func Tokenize(s string) []string {
	var toks []string
	for _, word := range Words(s) {
		toks = append(toks, TokenizeWord(word)...)
	}
	return toks
}

func Words(s string) []string {
	return sepRe.Split(s, -1)
}
