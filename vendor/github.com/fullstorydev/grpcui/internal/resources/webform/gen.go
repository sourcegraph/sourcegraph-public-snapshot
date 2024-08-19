package webform

//go:generate go-bindata -o=bindata.go -pkg=webform -nometadata webform-template.html webform-sample.css webform.js
//go:generate gofmt -w -s bindata.go
