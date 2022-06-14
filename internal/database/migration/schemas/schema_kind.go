package schemas

type SchemaKind interface {
	schemaKind()
}

type Frontend interface {
	SchemaKind
	frontend()
}

type CodeIntel interface {
	SchemaKind
	codeIntel()
}

type CodeInsights interface {
	SchemaKind
	codeInsights()
}

type Production interface {
	SchemaKind
	Frontend
	CodeIntel
}

type Any interface {
	SchemaKind
}

func SchemasFromKind(kind SchemaKind) (res []*Schema) {
	if _, ok := kind.(Frontend); ok {
		res = append(res, FrontendDefinition)
	}

	if _, ok := kind.(CodeIntel); ok {
		res = append(res, CodeIntelDefinition)
	}

	if _, ok := kind.(CodeInsights); ok {
		res = append(res, CodeInsightsDefinition)
	}

	return res
}
