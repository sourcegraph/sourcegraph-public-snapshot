import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { JsonDocument, JsonOccurrence, SyntaxKind } from '../../../lsif/lsif-typed'
import { BlobInfo } from '../Blob'

/**
 * This data structure combines the syntax highlighting data received from the
 * server with a lineIndex map (implemented as array), for fast lookup by line
 * number, with minimal additional impact on memory (e.g. garbage collection).
 */
interface HighlightIndex {
    occurrences: JsonOccurrence[]
    lineIndex: (number | undefined)[]
}

/**
 * Parses JSON-encoded SCIP syntax highlighting data and creates a line index.
 * NOTE: This assumes that the data is sorted and does not contain overlapping
 * ranges.
 */
function createHighlightTable(json: string | undefined): HighlightIndex {
    const lineIndex: (number | undefined)[] = []

    if (!json) {
        return { occurrences: [], lineIndex }
    }

    try {
        const occurrences = (JSON.parse(json) as JsonDocument).occurrences ?? []
        let previousEndline: number | undefined

        for (let index = 0; index < occurrences.length; index++) {
            const current = occurrences[index]
            const startLine = current.range[0]
            const endLine = current.range.length === 3 ? startLine : current.range[2]

            if (previousEndline !== startLine) {
                // Only use the current index if there isn't already an occurence on
                // the current line.
                lineIndex[startLine] = index
            }

            if (startLine !== endLine) {
                lineIndex[endLine] = index
            }

            previousEndline = endLine
        }

        return { occurrences, lineIndex }
    } catch {
        console.error(`Unable to process SCIP highlight data: ${json}`)
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

        const builder = new RangeSetBuilder<Decoration>()

        // Cache current line object
        let line = fromLine

        if (startIndex !== undefined) {
            // Iterate over the rendered line (numbers) and get the
            // corresponding occurrences from the highlighting table.
            for (let index = startIndex; index < occurrences.length; index++) {
                const occurrence = occurrences[index]

                if (occurrence.range[0] > toLine.number) {
                    break
                }

                if (occurrence.syntaxKind === undefined) {
                    continue
                }

                // Fetch new line information if necessary
                if (line.number !== occurrence.range[0] + 1) {
                    line = view.state.doc.line(occurrence.range[0] + 1)
                }

                builder.add(
                    line.from + occurrence.range[1],
                    occurrence.range.length === 3
                        ? line.from + occurrence.range[2]
                        : view.state.doc.line(occurrence.range[2] + 1).from + occurrence.range[3],
                    this.decorationCache[occurrence.syntaxKind] ||
                        (this.decorationCache[occurrence.syntaxKind] = Decoration.mark({
                            class: `hl-typed-${SyntaxKind[occurrence.syntaxKind]}`,
                        }))
                )
            }
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
        blobInfos[0]?.lsif ? createHighlightTable(blobInfos[0].lsif) : { occurrences: [], lineIndex: [] },
    enables: ViewPlugin.fromClass(SyntaxHighlightManager, { decorations: plugin => plugin.decorations }),
})
