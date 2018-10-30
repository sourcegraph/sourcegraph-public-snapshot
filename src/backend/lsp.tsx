import { AbsoluteRepo, FileSpec, PositionSpec } from '../repo'
import { HoverMerged, ModeSpec } from './features'

/**
 * Contains the fields necessary to route the request to the correct logical LSP server process and construct the
 * correct initialization request.
 */
export interface LSPSelector extends AbsoluteRepo, ModeSpec {}

/**
 * Contains the fields necessary to construct an LSP TextDocumentPositionParams value.
 */
export interface LSPTextDocumentPositionParams extends LSPSelector, PositionSpec, FileSpec {}

export const isEmptyHover = (hover: HoverMerged | null): boolean =>
    !hover ||
    !hover.contents ||
    (Array.isArray(hover.contents) && hover.contents.length === 0) ||
    hover.contents.every(c => {
        if (typeof c === 'object' && 'kind' in c && 'value' in c) {
            return !c.value
        }
        if (typeof c === 'string') {
            return !c
        }
        return !c.value
    })
