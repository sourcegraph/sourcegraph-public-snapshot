package render

import (
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/bevzzz/nb/render/internal/wildcard"
	"github.com/bevzzz/nb/schema"
)

// Renderer renders a decoded notebook in the format it implements.
type Renderer interface {
	// Render writes the contents of the notebook cells it supports.
	//
	// Implementations should not error on cell types, for which no RenderCellFunc is registered.
	// This is expected, as some [RawCells] will be rendered in some output formats and ignored in others.
	//
	// [RawCells]: https://nbformat.readthedocs.io/en/latest/format_description.html#raw-nbconvert-cells
	Render(io.Writer, schema.Notebook) error

	// AddOptions configures the editor after it has been constructured.
	// The renderer's configuration should not change between renders, and so, implementations should
	// ignore options added after the first call to Render().
	AddOptions(...Option)
}

// CellRenderer registers a RenderCellFunc for every cell type it supports.
//
// Reminiscent of the [Visitor] pattern, it allows extending the base renderer
// to support any number of arbitrary cell types.
//
// [Visitor]: https://refactoring.guru/design-patterns/visitor
type CellRenderer interface {
	// RegisterFuncs registers one or more RenderCellFunc with the passed renderer.
	RegisterFuncs(RenderCellFuncRegistry)
}

// RenderCellFuncRegistry is an interface that extendable Renderers should implement.
type RenderCellFuncRegistry interface {
	// Register adds a RenderCellFunc and a Pref selector for it.
	Register(Pref, RenderCellFunc)
}

// RenderCellFunc writes contents of a specific cell type.
type RenderCellFunc func(io.Writer, schema.Cell) error

type Config struct {
	CellWrapper
	CellRenderers []CellRenderer
}

type Option func(*Config)

// WithCellRenderers adds support for other cell types to the base renderer.
// If a renderer implements CellWrapper, it will be used to wrap input and output cells.
// Only one cell wrapper can be configured, and so the last implementor will take precedence.
func WithCellRenderers(crs ...CellRenderer) Option {
	return func(cfg *Config) {
		for _, cr := range crs {
			cfg.CellRenderers = append(cfg.CellRenderers, cr)

			if cw, ok := cr.(CellWrapper); ok {
				cfg.CellWrapper = cw
			}
		}
	}
}

// CellWrapper renders common wrapping elements for every cell type.
type CellWrapper interface {
	// Wrap the entire cell.
	Wrap(io.Writer, schema.Cell, RenderCellFunc) error

	// WrapInput wraps input block.
	WrapInput(io.Writer, schema.Cell, RenderCellFunc) error

	// WrapOuput wraps output block (code cells).
	WrapOutput(io.Writer, schema.Outputter, RenderCellFunc) error

	// WrapAll wraps all cells in the notebook.
	// This method will be called once and will receive a function
	// to render the rest of the notebook.
	WrapAll(io.Writer, func(io.Writer) error) error
}

// renderer is a base Renderer implementation.
// It does not support any cell types out of the box and should be extended by the client using the available Options.
type renderer struct {
	once   sync.Once
	config Config

	cellWrapper        CellWrapper
	renderCellFuncsTmp map[Pref]RenderCellFunc // renderCellFuncsTmp holds intermediary preference entries.
	renderCellFuncs    prefs                   // renderCellFuncs is sorted and will only be modified once.
}

var _ RenderCellFuncRegistry = (*renderer)(nil)

// NewRenderer extends the base renderer with the passed options.
func NewRenderer(opts ...Option) Renderer {
	r := renderer{
		renderCellFuncsTmp: make(map[Pref]RenderCellFunc),
	}
	r.AddOptions(opts...)
	return &r
}

var _ Renderer = (*renderer)(nil)
var _ RenderCellFuncRegistry = (*renderer)(nil)

func (r *renderer) AddOptions(opts ...Option) {
	for _, opt := range opts {
		opt(&r.config)
	}
}

// Register registers a new RenderCellFunc with a preference selector.
//
// Any function registered with the same Pref will be overridden. All configurations
// should be done the first call to Render(), as later changes will have no effect.
func (r *renderer) Register(pref Pref, f RenderCellFunc) {
	r.renderCellFuncsTmp[pref] = f
}

func (r *renderer) init() {
	r.once.Do(func() {
		r.cellWrapper = r.config.CellWrapper
		for _, cr := range r.config.CellRenderers {
			cr.RegisterFuncs(r)
		}
		for p, rf := range r.renderCellFuncsTmp {
			r.renderCellFuncs = append(r.renderCellFuncs, pref{
				Pref:   p,
				Render: rf,
			})
		}
		r.renderCellFuncs.Sort()
	})
}

// render renders the cell with the most-preferred RenderCellFunc.
//
// TODO: use sort.Find? need to try it out, like, because we have a mixed slice, where s[i] > s[i-1] might be true, but then s[i] and s[i-2] are semantically unrelated.
// Definitely not sort.Search, because sort.Search assumes that all elements >=i satisfy the condition, which is not the case.
func (r *renderer) render(w io.Writer, cell schema.Cell) error {
	for _, pref := range r.renderCellFuncs {
		if !pref.Match(cell) {
			continue
		}
		if err := pref.Render(w, cell); err != nil {
			// We could implement a failover mechanism, where, if the first-preference render fails,
			// we move on to the next matching option. The trouble here is that the first renderer
			// couldn've already written to io.Writer and we might end up with a corrupted document.
			//
			// Using an intermediate buffer buf and copying from it to w on successful render is an option,
			// but it adds some overhead and I wouldn't take it without a compelling case for this feature.
			return fmt.Errorf("nb: render: %w", err)
		}
		return nil
	}
	// TODO: currently we silently drop cells for which no render func is registered. Should we error?
	return nil
}

func (r *renderer) Render(w io.Writer, nb schema.Notebook) error {
	r.init()

	for _, cell := range nb.Cells() {
		var err error

		// TODO: lookup RenderCellFunc before opening the wrapper?

		if r.cellWrapper != nil {
			err = r.cellWrapper.Wrap(w, cell, func(w io.Writer, c schema.Cell) error {
				if err := r.cellWrapper.WrapInput(w, cell, r.render); err != nil {
					return err
				}

				if out, ok := cell.(interface{ schema.Outputter }); ok {
					if err := r.cellWrapper.WrapOutput(w, out, r.render); err != nil {
						return err
					}
				}
				return nil
			})
		} else {
			err = r.render(w, cell)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// Pref describes target cell and mime- type.
//
// Preference API is a flexible model which allows multiple CellRenderers
// to assume responsibility for specific cells. For example:
//
//	// Default renderer handles all "display_data" outputs: media, JSON, raw HTML, etc.
//	reg.Register(render.Pref{Type: schema.DisplayData}, r.renderDisplayData)
//
//	// This custom renderer only renders GIFs (regardless of the cell type).
//	reg.Register(render.Pref{MimeType: "image/gif"}, r.renderGIF)
//
//	// Finally, this renderer renders any other image media, but only from "display_data" outputs.
//	reg.Register(render.Pref{Type: schema.DisplayData, MimeType: "image/*"}, r.renderSQL)
//
// To provide this granularity, registered Prefs are sorted according to their:
//  1. Specificity: a measure for how precise the selection of target cells is.
//     Simply put, Type < MimeType < (Type+MimeType).
//  2. Wildcard count: Prefs with less "*" in their MimeType will be prioritized.
type Pref struct {
	// Type matches cells with the same Type().
	Type schema.CellType

	// MimeType matches cells based on their reported MimeType().
	// Use wildcard syntax (e.g. "image/*" or "*/*") to target
	// wider ranges of cell mime-types.
	MimeType string
}

// Match checks if the cell matches Pref's criteria.
func (p Pref) Match(cell schema.Cell) bool {
	if p.Type > schema.Unrecognized && p.Type != cell.Type() {
		return false
	}
	if p.MimeType != "" && !wildcard.Match(p.MimeType, cell.MimeType()) {
		return false
	}
	return true
}

// specificity calculates a score for how precise the selection of target cells is.
// Generally, Prefs that define more fields achieve greater specificity.
// Below are some examples:
//   - Type 			- (1)
//   - MimeType 		- (2)
//   - MimeType + Type	- (3)
//
// A larger increment is used to make sure MimeType yields greater value than Type.
// The exact values should not be relied upon, as they may change in the future.
func (p Pref) specificity() (s int) {
	if p.Type > schema.Unrecognized {
		s++
	}
	if p.MimeType != "" {
		s += 2
	}
	return
}

// pref adds RenderCellFunc to Pref to keep Pref hashable.
type pref struct {
	Pref
	Render RenderCellFunc
}

// prefs is a RenderCellFunc collection that sorts in the order of descending Pref specificity.
type prefs []pref

var _ sort.Interface = (*prefs)(nil)

// Sort preferences from most specific to least specific.
func (s prefs) Sort() {
	sort.Sort(s)
}

// Len is the number of pref elements.
func (s prefs) Len() int {
	return len(s)
}

// Swap swaps 2 pref elements.
func (s prefs) Swap(i, j int) {
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}

// Less returns true if s[i] is more specific than s[j].
func (s prefs) Less(i, j int) bool {
	return less(s[i].Pref, s[j].Pref)
}

// less returns true if p is more specific than other. It can be used to sort
// a slice of Prefs in the order of descending specificity:
//
//	sort.Slice(len(prefs), func(i, j int) bool {
//		return less(prefs[i], prefs[j])
//	})
//
// In addition to specificity, less considers mime-type semantics. That is, if both Prefs
// have non-zero MimeType and target the same Type (regardless which), less returns true
// if the other Pref uses more wildcards in its mime-type (which makes it less specific).
// For example, "text/*" is less specific than "text/plain", but more specific than "*/*".
func less(p, other Pref) bool {
	if s, sOther := p.specificity(), other.specificity(); s != sOther {
		return s > sOther
	}

	// Prefs that target different cell types are unrelated and can be sorted in any order.
	if p.Type != other.Type {
		return false
	}

	// At this point we know both Prefs have a non-zero MimeType,
	// otherwise their specificities would not be the same. Given that,
	// p must sort before other iff its MimeType is more exact (uses less wildcards).
	return wildcard.Count(p.MimeType) < wildcard.Count(other.MimeType)
}
