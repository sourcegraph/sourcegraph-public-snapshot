package schemas

type SchemaKind interface {
	schemaKind()
}

type Frontend struct {
	SchemaKind
}

func (Frontend) frontend() {}

type CodeIntel struct {
	SchemaKind
}

func (CodeIntel) codeIntel() {}

type CodeInsights struct {
	SchemaKind
}

func (CodeInsights) codeInsights() {}

type Production struct {
	SchemaKind
	Frontend
	CodeIntel
}

type Any struct {
	SchemaKind
}

func SchemasFromKind[T SchemaKind]() (res []*Schema) {
	var kind T
	if _, ok := any(kind).(interface{ frontend() }); ok {
		res = append(res, FrontendDefinition)
	}

	if _, ok := any(kind).(interface{ codeIntel() }); ok {
		res = append(res, CodeIntelDefinition)
	}

	if _, ok := any(kind).(interface{ codeInsights() }); ok {
		res = append(res, CodeInsightsDefinition)
	}

	return res
}
