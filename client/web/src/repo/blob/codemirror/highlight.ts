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

import { OccurrenceIndex } from './codeintel/occurrences'
import { positionToOffset } from './utils'

interface HighlightIndex {
    // allOccurrences is an index including all occurrences in the syntax highlighting data.
    allOccurrences: OccurrenceIndex
    // interactiveOccurrences is the interactive subset of allOccurrences in the syntax highlighting data.
    interactiveOccurrences: OccurrenceIndex
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
    let occurrences: Occurrence[] = []
    if (info.lsif) {
        try {
            occurrences = Occurrence.fromInfo(info)
        } catch (error) {
            logger.error(`Unable to process SCIP highlight data: ${info.lsif}`, error)
        }
    }
    return {
        allOccurrences: new OccurrenceIndex(occurrences),
        interactiveOccurrences: new OccurrenceIndex(occurrences.filter(isInteractiveOccurrence)),
    }
}

/**
 * Occurrences that are possibly interactive (i.e. they can have code intelligence).
 */
const INTERACTIVE_OCCURRENCE_KINDS = new Set([
    SyntaxKind.Identifier,
    SyntaxKind.IdentifierBuiltin,
    SyntaxKind.IdentifierConstant,
    SyntaxKind.IdentifierMutableGlobal,
    SyntaxKind.IdentifierParameter,
    SyntaxKind.IdentifierLocal,
    SyntaxKind.IdentifierShadowed,
    SyntaxKind.IdentifierModule,
    SyntaxKind.IdentifierFunction,
    SyntaxKind.IdentifierFunctionDefinition,
    SyntaxKind.IdentifierMacro,
    SyntaxKind.IdentifierMacroDefinition,
    SyntaxKind.IdentifierType,
    SyntaxKind.IdentifierAttribute,
])

function isInteractiveOccurrence(occurrence: Occurrence): boolean {
    if (!occurrence.kind) {
        return false
    }

    return INTERACTIVE_OCCURRENCE_KINDS.has(occurrence.kind)
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

            const occurrences = view.state.facet(syntaxHighlight).allOccurrences

            // Find index of first relevant token
            let startIndex: number | undefined
            {
                let line = fromLine.number - 1
                do {
                    startIndex = occurrences.lineIndex[line++]
                } while (startIndex === undefined && line < occurrences.lineIndex.length)
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
    combine: values => createHighlightTable(values[0] ?? {}),
    enables: ViewPlugin.fromClass(SyntaxHighlightManager, { decorations: plugin => plugin.decorations }),
})
