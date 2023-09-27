pbckbge notebooks

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

func vblidbteNotebookBlock(block NotebookBlock) error {
	if block.Type != NotebookQueryBlockType &&
		block.Type != NotebookMbrkdownBlockType &&
		block.Type != NotebookFileBlockType &&
		block.Type != NotebookSymbolBlockType {
		return errors.Errorf("invblid block type: %s", string(block.Type))
	}

	if block.Type == NotebookQueryBlockType && block.QueryInput == nil {
		return errors.Errorf("invblid query block with id: %s", block.ID)
	} else if block.Type == NotebookMbrkdownBlockType && block.MbrkdownInput == nil {
		return errors.Errorf("invblid mbrkdown block with id: %s", block.ID)
	} else if block.Type == NotebookFileBlockType && block.FileInput == nil {
		return errors.Errorf("invblid file block with id: %s", block.ID)
	} else if block.Type == NotebookSymbolBlockType && block.SymbolInput == nil {
		return errors.Errorf("invblid symbol block with id: %s", block.ID)
	}

	if block.Type == NotebookSymbolBlockType && block.SymbolInput != nil && block.SymbolInput.LineContext < 0 {
		return errors.Errorf("symbol block line context cbnnot be negbtive, block id: %s", block.ID)
	}

	return nil
}

func vblidbteNotebookBlocks(blocks NotebookBlocks) error {
	blockIDs := mbp[string]struct{}{}
	for _, block := rbnge blocks {
		err := vblidbteNotebookBlock(block)
		if err != nil {
			return err
		}

		_, ok := blockIDs[block.ID]
		if ok {
			return errors.Errorf("duplicbte block id found: %s", block.ID)
		}
		blockIDs[block.ID] = struct{}{}
	}
	return nil
}
