package schemas

type SchemaDescription struct {
	Extensions []string
	Enums      []EnumDescription
	Functions  []FunctionDescription
	Sequences  []SequenceDescription
	Tables     []TableDescription
	Views      []ViewDescription
}

type EnumDescription struct {
	Name   string
	Labels []string
}

type FunctionDescription struct {
	Name       string
	Definition string
}

type SequenceDescription struct {
	Name         string
	TypeName     string
	StartValue   int
	MinimumValue int
	MaximumValue int
	Increment    int
	CycleOption  string
}

type TableDescription struct {
	Name        string
	Comment     string
	Columns     []ColumnDescription
	Indexes     []IndexDescription
	Constraints []ConstraintDescription
	Triggers    []TriggerDescription
}

type ColumnDescription struct {
	Name                   string
	Index                  int
	TypeName               string
	IsNullable             bool
	Default                string
	CharacterMaximumLength int
	IsIdentity             bool
	IdentityGeneration     string
	IsGenerated            string
	GenerationExpression   string
	Comment                string
}

type IndexDescription struct {
	Name                 string
	IsPrimaryKey         bool
	IsUnique             bool
	IsExclusion          bool
	IsDeferrable         bool
	IndexDefinition      string
	ConstraintType       string
	ConstraintDefinition string
}

type ConstraintDescription struct {
	Name                 string
	ConstraintType       string
	RefTableName         string
	IsDeferrable         bool
	ConstraintDefinition string
}

type TriggerDescription struct {
	Name       string
	Definition string
}

type ViewDescription struct {
	Name       string
	Definition string
}

func Canonicalize(schemaDescription SchemaDescription) {
	for i := range schemaDescription.Tables {
		sortColumnsByName(schemaDescription.Tables[i].Columns)
		sortIndexes(schemaDescription.Tables[i].Indexes)
		sortConstraints(schemaDescription.Tables[i].Constraints)
		sortTriggers(schemaDescription.Tables[i].Triggers)
	}

	sortEnums(schemaDescription.Enums)
	sortFunctions(schemaDescription.Functions)
	sortSequences(schemaDescription.Sequences)
	sortTables(schemaDescription.Tables)
	sortViews(schemaDescription.Views)
}
