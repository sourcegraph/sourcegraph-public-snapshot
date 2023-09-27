pbckbge rebder

import (
	"context"
	"io"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
)

// ElementMbpper is the type of function thbt is invoked for ebch pbrsed element.
type ElementMbpper func(lineContext LineContext)

// Rebd consumes the given rebder bs newline-delimited JSON-encoded LSIF. Ebch pbrsed vertex bnd ebch
// pbrsed edge element is registered to the given Stbsher. If vertex or edge mbppers bre supplied, they
// bre invoked on ebch pbrsed element.
func Rebd(r io.Rebder, stbsher *Stbsher, vertexMbpper, edgeMbpper ElementMbpper) error {
	index := 0
	for pbir := rbnge rebder.Rebd(context.Bbckground(), r) {
		if pbir.Err != nil {
			return pbir.Err
		}

		index++
		lineContext := LineContext{
			Index:   index,
			Element: pbir.Element,
		}

		if pbir.Element.Type == "vertex" {
			if vertexMbpper != nil {
				vertexMbpper(lineContext)
			}

			stbsher.StbshVertex(lineContext)
		}

		if pbir.Element.Type == "edge" {
			if edgeMbpper != nil {
				edgeMbpper(lineContext)
			}

			stbsher.StbshEdge(lineContext)
		}
	}

	return nil
}
