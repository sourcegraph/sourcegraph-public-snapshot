import { Facet, RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    type DecorationSet,
    type EditorView,
    type PluginValue,
    ViewPlugin,
    type ViewUpdate,
} from '@codemirror/view'

import { logger } from '@sourcegraph/common'
import { Occurrence, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { positionToOffset } from './utils'

/**
 * This data structure combines the syntax highlighting data received from the
 * server with a lineIndex map (implemented as array), for fast lookup by line
 * number, with minimal additional impact on memory (e.g. garbage collection).
 */
export interface HighlightIndex {
    occurrences: Occurrence[]
    lineIndex: (number | undefined)[]
}

interface HighlightData {
    content: string
    lsif?: string
}

/**
 * Parses JSON-encoded SCIP syntax highlighting data and creates a line index.
 * NOTE: This assumes that the data is sorted and does not contain overlapping
 * ranges.
 */
export function createHighlightTable(info: HighlightData): HighlightIndex {
    const lineIndex: (number | undefined)[] = []

    if (!info.lsif) {
        return { occurrences: [], lineIndex }
    }

    try {
        const occurrences = Occurrence.fromInfo(info)
        let previousEndline: number | undefined

        for (let index = 0; index < occurrences.length; index++) {
            const current = occurrences[index]

            if (previousEndline !== current.range.start.line) {
                // Only use the current index if there isn't already an occurrence on
                // the current line.
                lineIndex[current.range.start.line] = index
            }

            if (!current.range.isSingleLine()) {
                lineIndex[current.range.end.line] = index
            }

            previousEndline = current.range.end.line
        }

        return { occurrences, lineIndex }
    } catch (error) {
        logger.error(`Unable to process SCIP highlight data: ${info.lsif}`, error)
        return { occurrences: [], lineIndex }
    }
}

class SyntaxHighlightManager implements PluginValue {
    private decorationCache: Partial<Record<SyntaxKind, Decoration>> = {}
    public decorations: DecorationSet = Decoration.none

    constructor(view: EditorView) {
        this.decorations = this.computeDecorations(view)
    }

    public update(update: ViewUpdate): void {
        if (update.viewportChanged) {
            this.decorations = this.computeDecorations(update.view)
        }
    }

    private computeDecorations(view: EditorView): DecorationSet {
        const builder = new RangeSetBuilder<Decoration>()
        try {
            const { from, to } = view.viewport

            // Determine the start and end lines of the current viewport
            const fromLine = view.state.doc.lineAt(from)
            const toLine = view.state.doc.lineAt(to)

            const { occurrences, lineIndex } = view.state.facet(syntaxHighlight)

            // Find index of first relevant token
            let startIndex: number | undefined
            {
                let line = fromLine.number - 1
                do {
                    startIndex = lineIndex[line++]
                } while (startIndex === undefined && line < lineIndex.length)
            }

            // Cache current line object
            let line = fromLine

            if (startIndex !== undefined) {
                // Iterate over the rendered line (numbers) and get the
                // corresponding occurrences from the highlighting table.
                const textDocument = view.state.doc

                for (let index = startIndex; index < occurrences.length; index++) {
                    const occurrence = occurrences[index]
                    const occurrenceStartLine = occurrence.range.start.line + 1

                    if (occurrenceStartLine > toLine.number) {
                        break
                    }

                    if (occurrence.kind === undefined) {
                        continue
                    }

                    // Fetch new line information if necessary
                    if (line.number !== occurrenceStartLine) {
                        // If the next occurrence doesn't map to a valid
                        // document position, stop
                        if (occurrenceStartLine > textDocument.lines) {
                            break
                        }
                        line = textDocument.line(occurrenceStartLine)
                    }

                    const from = Math.min(line.from + occurrence.range.start.character, line.to)
                    // Should the range end be not a valid position in the
                    // document we fall back to the end of the current line
                    const to = occurrence.range.isSingleLine()
                        ? Math.min(line.from + occurrence.range.end.character, line.to)
                        : positionToOffset(textDocument, occurrence.range.end) ?? line.to

                    const decoration =
                        this.decorationCache[occurrence.kind] ||
                        (this.decorationCache[occurrence.kind] = Decoration.mark({
                            class: `hl-typed-${SyntaxKind[occurrence.kind]}`,
                        }))
                    builder.add(from, to, decoration)
                }
            }
        } catch (error) {
            logger.error('Failed to compute decorations from SCIP occurrences', error)
        }
        return builder.finish()
    }
}

/**
 * Facet for providing syntax highlighting information from a {@link BlobInfo}
 * object.
 */
export const syntaxHighlight = Facet.define<HighlightData, HighlightIndex>({
    static: true,
    compareInput: (inputA, inputB) => inputA.lsif === inputB.lsif,
    combine: values => (values[0]?.lsif ? createHighlightTable(values[0]) : { occurrences: [], lineIndex: [] }),
    enables: ViewPlugin.fromClass(SyntaxHighlightManager, { decorations: plugin => plugin.decorations }),
})
