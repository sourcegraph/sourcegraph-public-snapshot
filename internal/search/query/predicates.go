package query

// All recognized predicates
var AllPredicates = map[string]map[string]struct{}{
	FieldRepo: {
		"contains": empty,
	},
}
