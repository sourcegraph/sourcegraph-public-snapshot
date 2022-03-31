package schemas

type SchemaFormatter interface {
	Format(schemaDescription SchemaDescription) string
}
