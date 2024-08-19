package ast

import "fmt"

// LoopUp allows for circular references to be used.
type LoopUp struct {
	Key   string
	Table *map[string]interface{}
}

func (l *LoopUp) Get() (interface{}, error) {
	table := *l.Table
	i, ok := table[l.Key]
	if !ok {
		return nil, &LoopUpError{
			Value: *l,
		}
	}
	return i, nil
}

// LoopUpError is an error that occurs when a key can not be found in the table.
type LoopUpError struct {
	Value LoopUp
}

func (e *LoopUpError) Error() string {
	return fmt.Sprintf("loopup: can not find %s in table", e.Value.Key)
}
