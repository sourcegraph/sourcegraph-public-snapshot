package notebooks

import "github.com/sourcegraph/sourcegraph/lib/errors"

func validateNotebookBlock(block NotebookBlock) error {
	if block.Type != NotebookQueryBlockType && block.Type != NotebookMarkdownBlockType && block.Type != NotebookFileBlockType {
		return errors.Errorf("invalid block type: %s", string(block.Type))
	}

	if block.Type == NotebookQueryBlockType && block.QueryInput == nil {
		return errors.Errorf("invalid query block with id: %s", block.ID)
	} else if block.Type == NotebookMarkdownBlockType && block.MarkdownInput == nil {
		return errors.Errorf("invalid markdown block with id: %s", block.ID)
	} else if block.Type == NotebookFileBlockType && block.FileInput == nil {
		return errors.Errorf("invalid file block with id: %s", block.ID)
	}

	return nil
}

func validateNotebookBlocks(blocks NotebookBlocks) error {
	blockIDs := map[string]struct{}{}
	for _, block := range blocks {
		err := validateNotebookBlock(block)
		if err != nil {
			return err
		}

		_, ok := blockIDs[block.ID]
		if ok {
			return errors.Errorf("duplicate block id found: %s", block.ID)
		}
		blockIDs[block.ID] = struct{}{}
	}
	return nil
}
