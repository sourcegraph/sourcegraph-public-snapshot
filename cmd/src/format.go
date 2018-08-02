package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

func parseTemplate(text string) (*template.Template, error) {
	tmpl := template.New("")
	tmpl.Funcs(map[string]interface{}{
		"json": func(v interface{}) (string, error) {
			b, err := json.MarshalIndent(v, "", "  ")
			return string(b), err
		},
	})
	return tmpl.Parse(text)
}

func execTemplate(tmpl *template.Template, data interface{}) error {
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// json.MarshalIndent, but with defaults.
func marshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
