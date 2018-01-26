package pkg

//go:cgo_export_dynamic
func foo() {}

func bar() {} // MATCH /bar is unused/
