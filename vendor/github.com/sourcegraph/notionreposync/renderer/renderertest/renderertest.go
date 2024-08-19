package renderertest

import (
	"context"

	"github.com/jomei/notionapi"
)

// MockBlockUpdater is a mock BlockUpdater that records the blocks added to it.
// Recorded blocks can be collected with 'GetAddedBlocks()'.
type MockBlockUpdater struct{ addedBlocks []notionapi.Block }

func (m *MockBlockUpdater) AddChildren(ctx context.Context, children []notionapi.Block) error {
	m.addedBlocks = append(m.addedBlocks, children...)
	return nil
}

func (m *MockBlockUpdater) GetAddedBlocks() []notionapi.Block {
	return m.addedBlocks
}
