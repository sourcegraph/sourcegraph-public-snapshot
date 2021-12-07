package schemas

var SchemaNames []string

func init() {
	for _, schema := range Schemas {
		SchemaNames = append(SchemaNames, schema.Name)
	}
}
