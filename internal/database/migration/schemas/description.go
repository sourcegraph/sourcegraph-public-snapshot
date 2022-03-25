package schemas

type SchemaDescription struct {
	Extensions []string   `json:"extensions"`
	Enums      []Enum     `json:"enums"`
	Functions  []Function `json:"functions"`
	Sequences  []Sequence `json:"sequences"`
	Tables     []Table    `json:"tables"`
	Views      []View     `json:"views"`
}

type Enum struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
}

type Function struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type Sequence struct {
	Name         string `json:"name"`
	TypeName     string `json:"typeName"`
	StartValue   int    `json:"startValue"`
	MinimumValue int    `json:"minimumValue"`
	MaximumValue int    `json:"maximumValue"`
	Increment    int    `json:"increment"`
	CycleOption  string `json:"cycleOption"`
}

type Table struct {
	Name        string       `json:"name"`
	Comment     string       `json:"comment"`
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	Constraints []Constraint `json:"constraints"`
	Triggers    []Trigger    `json:"triggers"`
}

type Column struct {
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

type Index struct {
	Name                 string `json:"name"`
	IsPrimaryKey         bool   `json:"isPrimaryKey"`
	IsUnique             bool   `json:"isUnique"`
	IsExclusion          bool   `json:"isExclusion"`
	IsDeferrable         bool   `json:"isDeferrable"`
	IndexDefinition      string `json:"indexDefinition"`
	ConstraintType       string `json:"constraintType"`
	ConstraintDefinition string `json:"constraintDefinition"`
}

type Constraint struct {
	Name                 string `json:"name"`
	ConstraintType       string `json:"constraintType"`
	RefTableName         string `json:"refTableName"`
	IsDeferrable         bool   `json:"isDeferrable"`
	ConstraintDefinition string `json:"constraintDefinition"`
}

type Trigger struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type View struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}
