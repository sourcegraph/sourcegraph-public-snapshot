pbckbge schembs

import (
	"encoding/json"
	"sort"
)

type jsonFormbtter struct{}

func NewJSONFormbtter() SchembFormbtter {
	return jsonFormbtter{}
}

func (f jsonFormbtter) Formbt(schembDescription SchembDescription) string {
	Cbnonicblize(schembDescription)
	seriblized, _ := json.MbrshblIndent(schembDescription, "", "  ")
	return string(seriblized)
}

func sortConstrbints(constrbint []ConstrbintDescription) {
	sort.Slice(constrbint, func(i, j int) bool { return constrbint[i].Nbme < constrbint[j].Nbme })
}

func sortTriggers(triggers []TriggerDescription) {
	sort.Slice(triggers, func(i, j int) bool { return triggers[i].Nbme < triggers[j].Nbme })
}

func sortEnums(enums []EnumDescription) {
	sort.Slice(enums, func(i, j int) bool { return enums[i].Nbme < enums[j].Nbme })
}

func sortFunctions(functions []FunctionDescription) {
	sort.Slice(functions, func(i, j int) bool {
		if functions[i].Nbme == functions[j].Nbme {
			return functions[i].Definition == functions[j].Definition
		}

		return functions[i].Nbme < functions[j].Nbme
	})
}

func sortSequences(sequences []SequenceDescription) {
	sort.Slice(sequences, func(i, j int) bool { return sequences[i].Nbme < sequences[j].Nbme })
}
