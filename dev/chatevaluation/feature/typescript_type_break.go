package feature

import "regexp"

type TypeScriptTypeBreak struct{}

// Does not work that well, for instance will replace // TODO: foo with // TODO: string.
func (f TypeScriptTypeBreak) Distort(contents string) string {
	typeAnnotation := regexp.MustCompile(`:\s*([a-zA-Z\[\]<>.]+)`)
	matches := typeAnnotation.FindStringSubmatch(contents)
	if len(matches) > 0 {
		if matches[1] != ": string" {
			var replaced bool
			return typeAnnotation.ReplaceAllStringFunc(contents, func(typ string) string {
				if replaced {
					return typ
				}
				if typ != ": string" {
					replaced = true
				}
				return ": string"
			})
		}
	}
	return contents
}

func (f TypeScriptTypeBreak) ValidateFile(got, want string) bool {
	return got == want
}
