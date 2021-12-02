package notebooks

import (
	"encoding/json"
)

func blockInput(block NotebookBlock) interface{} {
	switch block.Type {
	case NotebookQueryBlockType:
		return block.QueryInput
	case NotebookMarkdownBlockType:
		return block.MarkdownInput
	case NotebookFileBlockType:
		return block.FileInput
	}
	panic("unhandled block input type")
}

func (block NotebookBlock) MarshalJSON() ([]byte, error) {
	type Alias NotebookBlock
	return json.Marshal(&struct {
		Alias
		Input interface{} `json:"input"`
	}{
		Alias: (Alias)(block),
		Input: blockInput(block),
	})
}

func (block *NotebookBlock) UnmarshalJSON(bytes []byte) error {
	type Alias NotebookBlock
	aux := &struct {
		*Alias
		Input interface{} `json:"input"`
	}{
		Alias: (*Alias)(block),
	}
	if err := json.Unmarshal(bytes, &aux); err != nil {
		return err
	}

	// We cannot cast aux.Input directly into the appropriate input type, because aux.Input
	// has the map[string]interface{} underlying type. We could manually extract the input fields
	// based on the block type, but that is cumbersome and error-prone. Instead, we marshal the input
	// again and unmarshal it into the appropriate input type.
	marshalledInput, err := json.Marshal(aux.Input)
	if err != nil {
		return err
	}

	switch block.Type {
	case NotebookQueryBlockType:
		if err = json.Unmarshal(marshalledInput, &block.QueryInput); err != nil {
			return err
		}
	case NotebookMarkdownBlockType:
		if err = json.Unmarshal(marshalledInput, &block.MarkdownInput); err != nil {
			return err
		}
	case NotebookFileBlockType:
		if err = json.Unmarshal(marshalledInput, &block.FileInput); err != nil {
			return err
		}
	}
	return nil
}
