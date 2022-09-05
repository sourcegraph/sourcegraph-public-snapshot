import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { Occurrence, SyntaxKind } from '../../../lsif/lsif-typed'
import { BlobInfo } from '../Blob'

/**
 * This data structure combines the syntax highlighting data received from the
 * server with a lineIndex map (implemented as array), for fast lookup by line
 * number, with minimal additional impact on memory (e.g. garbage collection).
 */
interface HighlightIndex {
    occurrences: Occurrence[]
    lineIndex: (number | undefined)[]
}

/**
 * Parses JSON-encoded SCIP syntax highlighting data and creates a line index.
 * NOTE: This assumes that the data is sorted and does not contain overlapping
 * ranges.
 */
function createHighlightTable(info: BlobInfo): HighlightIndex {
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
                // Only use the current index if there isn't already an occurence on
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
        console.error(`Unable to process SCIP highlight data: ${info.lsif}`, error)
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
                for (let index = startIndex; index < occurrences.length; index++) {
                    const occurrence = occurrences[index]

                    if (occurrence.range.start.line > toLine.number) {
                        break
                    }

                    if (occurrence.kind === undefined) {
                        continue
                    }

                    // Fetch new line information if necessary
                    if (line.number !== occurrence.range.start.line + 1) {
                        line = view.state.doc.line(occurrence.range.start.line + 1)
                    }

                    const from = line.from + occurrence.range.start.character
                    const to = occurrence.range.isSingleLine()
                        ? line.from + occurrence.range.end.character
                        : view.state.doc.line(occurrence.range.end.line + 1).from + occurrence.range.end.character
                    const decoration =
                        this.decorationCache[occurrence.kind] ||
                        (this.decorationCache[occurrence.kind] = Decoration.mark({
                            class: `hl-typed-${SyntaxKind[occurrence.kind]}`,
                        }))
                    builder.add(from, to, decoration)
                }
            }
        } catch (error) {
            console.error('Failed to compute decorations from SCIP occurrences', error)
        }
        return builder.finish()
    }
}

/**
 * Facet for providing syntax highlighting information from a {@link BlobInfo}
 * object.
 */
export const syntaxHighlight = Facet.define<BlobInfo, HighlightIndex>({
    static: true,
    compareInput: (blobInfoA, blobInfoB) => blobInfoA.lsif === blobInfoB.lsif,
    combine: blobInfos =>
        blobInfos[0]?.lsif ? createHighlightTable(blobInfos[0]) : { occurrences: [], lineIndex: [] },
    enables: ViewPlugin.fromClass(SyntaxHighlightManager, { decorations: plugin => plugin.decorations }),
})
