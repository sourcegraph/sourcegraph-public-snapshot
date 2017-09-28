package pkg

func fn(i interface{}) {
	if _, ok := i.(string); ok && i != nil { // MATCH "when ok is true, i can't be nil"
	}
	if _, ok := i.(string); i != nil && ok { // MATCH "when ok is true, i can't be nil"
	}
	if _, ok := i.(string); i != nil || ok {
	}
	if _, ok := i.(string); i != nil && !ok {
	}
	if _, ok := i.(string); i == nil && ok {
	}
}
