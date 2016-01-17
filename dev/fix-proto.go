package main

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
)

var linePattern = regexp.MustCompile(`^(\s+)(repeated )?((?:[\w\d\.]|<.*>)+ )?([\w\d]+)( = )(\d+)( \[(?:.+)\])?(;)( //.*)?(\n)$`)
var wordBeginPattern = regexp.MustCompile(`^\w|_\w`)
var optionPattern = regexp.MustCompile(`^\(([\w\.]+)\) = (?:(\w+)|"(.*)")$`)

func main() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	out, err := os.Create(os.Args[1] + ".new")
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if !strings.Contains(line, "=") || strings.Contains(line, "google.api.http") || strings.HasPrefix(strings.TrimSpace(line), "//") || strings.HasPrefix(line, "syntax") || strings.HasPrefix(line, "option") {
			if _, err := out.WriteString(line); err != nil {
				panic(err)
			}
			continue
		}

		matches := linePattern.FindStringSubmatch(line)
		if matches == nil {
			panic("line pattern did not match")
		}
		parts := matches[1:]

		parts[3] = wordBeginPattern.ReplaceAllStringFunc(parts[3], func(s string) string { return strings.ToUpper(strings.TrimPrefix(s, "_")) })

		if len(parts[6]) != 0 {
			options := strings.Split(parts[6][2:len(parts[6])-1], ", ")
			var newOptions []string
			hasJSONTag := false
			isEmbedded := false
			for _, o := range options {
				matches := optionPattern.FindStringSubmatch(o)
				if matches == nil {
					panic("option pattern did not match")
				}
				switch matches[1] {
				case "gogoproto.customname":
					parts[3] = matches[3]
					continue
				case "gogoproto.jsontag":
					hasJSONTag = true
				case "gogoproto.embed":
					if matches[2] == "true" {
						isEmbedded = true
					}
				}
				newOptions = append(newOptions, o)
			}
			if !hasJSONTag && isEmbedded {
				newOptions = append(newOptions, `(gogoproto.jsontag) = ""`)
			}
			parts[6] = " [" + strings.Join(newOptions, ", ") + "]"
			if len(newOptions) == 0 {
				parts[6] = ""
			}
		}

		if _, err := out.WriteString(strings.Join(parts, "")); err != nil {
			panic(err)
		}
	}

	in.Close()
	out.Close()
	os.Rename(os.Args[1]+".new", os.Args[1])
}
