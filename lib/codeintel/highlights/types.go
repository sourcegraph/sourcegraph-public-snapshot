package highlights

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EngineType int

const (
	EngineInvalid EngineType = iota
	EngineTreeSitter
	EngineSyntect
)

// NOTE: Do not change these values, they are used across application boundaries
// (for highlighting requests) so changing strings will cause problems in syntax-highlighter
// service.
//
// Check `pub enum SyntaxEngine` in syntax-highlighter for current values.
var engineToDisplay map[EngineType]string = map[EngineType]string{
	EngineInvalid:    "invalid",
	EngineSyntect:    "syntect",
	EngineTreeSitter: "tree-sitter",
}

func EngineTypeToNameChecked(engine EngineType) (string, error) {
	name, ok := engineToDisplay[engine]
	if engine == EngineInvalid || !ok {
		return "", errors.New("Invalid Engine Type")
	}

	return name, nil
}

func EngineTypeToName(engine EngineType) string {
	name, ok := engineToDisplay[engine]
	if !ok {
		return engineToDisplay[EngineInvalid]
	}

	return name
}

func EngineNameToEngineType(engineName string) (engine EngineType, ok bool) {
	switch engineName {
	case "tree-sitter":
		return EngineTreeSitter, true
	case "syntect":
		return EngineSyntect, true
	default:
		return EngineInvalid, false
	}
}
