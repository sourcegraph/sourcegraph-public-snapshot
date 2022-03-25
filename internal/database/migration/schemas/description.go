package schemas

type SchemaDescription struct {
	Extensions []string              `json:"extensions"`
	Enums      []EnumDescription     `json:"enums"`
	Functions  []FunctionDescription `json:"functions"`
	Sequences  []SequenceDescription `json:"sequences"`
	Tables     []TableDescription    `json:"tables"`
	Views      []ViewDescription     `json:"views"`
}

type EnumDescription struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
}

type FunctionDescription struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type SequenceDescription struct {
	Name         string `json:"name"`
	TypeName     string `json:"typeName"`
	StartValue   int    `json:"startValue"`
	MinimumValue int    `json:"minimumValue"`
	MaximumValue int    `json:"maximumValue"`
	Increment    int    `json:"increment"`
	CycleOption  string `json:"cycleOption"`
}

type TableDescription struct {
	Name        string                  `json:"name"`
	Comment     string                  `json:"comment"`
	Columns     []ColumnDescription     `json:"columns"`
	Indexes     []IndexDescription      `json:"indexes"`
	Constraints []ConstraintDescription `json:"constraints"`
	Triggers    []TriggerDescription    `json:"triggers"`
}

type ColumnDescription struct {
	Name                   string `json:"name"`
	Index                  int    `json:"index"`
	TypeName               string `json:"typeName"`
	IsNullable             bool   `json:"isNullable"`
	Default                string `json:"default"`
	CharacterMaximumLength int    `json:"characterMaximumLength"`
	IsIdentity             bool   `json:"isIdentity"`
	IdentityGeneration     string `json:"identityGeneration"`
	IsGenerated            string `json:"isGenerated"`
	GenerationExpression   string `json:"generationExpression"`
	Comment                string `json:"comment"`
}

type IndexDescription struct {
	Name                 string `json:"name"`
	IsPrimaryKey         bool   `json:"isPrimaryKey"`
	IsUnique             bool   `json:"isUnique"`
	IsExclusion          bool   `json:"isExclusion"`
	IsDeferrable         bool   `json:"isDeferrable"`
	IndexDefinition      string `json:"indexDefinition"`
	ConstraintType       string `json:"constraintType"`
	ConstraintDefinition string `json:"constraintDefinition"`
}

type ConstraintDescription struct {
	Name                 string `json:"name"`
	ConstraintType       string `json:"constraintType"`
	RefTableName         string `json:"refTableName"`
	IsDeferrable         bool   `json:"isDeferrable"`
	ConstraintDefinition string `json:"constraintDefinition"`
}

type TriggerDescription struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type ViewDescription struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}
