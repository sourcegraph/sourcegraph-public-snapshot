package schemas

import (
	"encoding/json"
	"sort"
)

type jsonFormatter struct{}

func NewJSONFormatter() SchemaFormatter {
	return jsonFormatter{}
}

func (f jsonFormatter) Format(schemaDescription SchemaDescription) string {
	Canonicalize(schemaDescription)
	serialized, _ := json.MarshalIndent(schemaDescription, "", "  ")
	return string(serialized)
}

func sortConstraints(constraint []ConstraintDescription) {
	sort.Slice(constraint, func(i, j int) bool { return constraint[i].Name < constraint[j].Name })
}

func sortTriggers(triggers []TriggerDescription) {
	sort.Slice(triggers, func(i, j int) bool { return triggers[i].Name < triggers[j].Name })
}

func sortEnums(enums []EnumDescription) {
	sort.Slice(enums, func(i, j int) bool { return enums[i].Name < enums[j].Name })
}

func sortFunctions(functions []FunctionDescription) {
	sort.Slice(functions, func(i, j int) bool {
		if functions[i].Name == functions[j].Name {
			return functions[i].Definition == functions[j].Definition
		}

		return functions[i].Name < functions[j].Name
	})
}

func sortSequences(sequences []SequenceDescription) {
	sort.Slice(sequences, func(i, j int) bool { return sequences[i].Name < sequences[j].Name })
}
