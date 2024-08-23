package decode

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

// Bytes decodes raw JSON bytes into a version-specific notebook representation.
func Bytes(b []byte) (schema.Notebook, error) {
	var nb notebook
	if err := json.Unmarshal(b, &nb); err != nil {
		return nil, fmt.Errorf("decode: bytes: %w", err)
	}
	return &nb, nil
}

// notebook unmarshals raw .ipynb (JSON) to the right schema based on its version.
type notebook struct {
	common.Notebook
	cells []schema.Cell
}

func (n *notebook) Cells() []schema.Cell {
	return n.cells
}

func (n *notebook) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &n.Notebook); err != nil {
		return err
	}

	ver := n.Version()
	d, ok := getDecoder(ver)
	if !ok {
		return fmt.Errorf("schema %s is not supported", ver)
	}

	meta, err := d.DecodeMeta(n.Notebook.Metadata)
	if err != nil {
		return fmt.Errorf("%s: notebook metadata: %w", ver, err)
	}

	cells, err := d.ExtractCells(data)
	if err != nil {
		return fmt.Errorf("%s: extract cells: %w", ver, err)
	}

	n.cells = make([]schema.Cell, len(cells))
	for i, raw := range cells {
		c := cell{meta: meta, decoder: d}
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("%s: %w", ver, err)
		}
		n.cells[i] = c.cell
	}
	return nil
}

// cell unmarshals cell bytes according to the notebook version.
type cell struct {
	decoder Decoder
	meta    schema.NotebookMetadata
	cell    schema.Cell
}

func (c *cell) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	cell, err := c.decoder.DecodeCell(v, data, c.meta)
	if err != nil {
		return fmt.Errorf("cell content: %w", err)
	}
	c.cell = cell
	return nil
}

// Decoder implementations are version-aware and decode cell contents and metadata
// based on the respective JSON schema definition.
type Decoder interface {
	// ExtractCells accesses the array of notebook cells.
	//
	// Prior to v4.0 cells were not a part of the top level structure,
	// and were contained in "worksheets" instead.
	ExtractCells(data []byte) ([]json.RawMessage, error)

	// DecodeMeta decodes version-specific metadata.
	DecodeMeta(data []byte) (schema.NotebookMetadata, error)

	// DecodeCell decodes raw cell data to a version-specific implementation.
	DecodeCell(v map[string]interface{}, data []byte, meta schema.NotebookMetadata) (schema.Cell, error)
}

var (
	mutex    = new(sync.RWMutex)
	decoders = make(map[schema.Version]Decoder)
)

// RegisterDecoder for a schema version.
func RegisterDecoder(v schema.Version, d Decoder) {
	mutex.Lock()
	defer mutex.Unlock()
	if d == nil {
		panic("invalid nil decoder")
	}
	decoders[v] = d
}

// getDecoder by schema version.
func getDecoder(v schema.Version) (d Decoder, ok bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	d, ok = decoders[v]
	return
}
