import { EditorSelection, type Text, type EditorState, type SelectionRange } from '@codemirror/state'

import type { Range } from '@sourcegraph/extension-api-types'
import { Occurrence, Position, Range as ScipRange, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { codeGraphData } from './codeintel/occurrences'
import { syntaxHighlight } from './highlight'

export function interactiveOccurrenceAt(state: EditorState, offset: number): Occurrence | undefined {
    const position = positionAtCmPosition(state.doc, offset)

    // First we try to get an occurrence from the occurrences API
    const data = state.facet(codeGraphData)
    if (data.length > 0) {
        // Arbitrarily choose the first set of code graph data
        // because we have no good heuristics for selecting between
        // multiple.
        return data[0].occurrenceIndex.atPosition(position)
    }

    // Next we try to get an occurrence from syntax highlighting data.
    const highlightingOccurrences = state.facet(syntaxHighlight).interactiveOccurrences
    if (highlightingOccurrences.length > 0) {
        return highlightingOccurrences.atPosition(position)
    }

    // If the syntax highlighting data is incomplete then we fallback to a
    // heursitic to infer the occurrence.
    return inferOccurrenceAtOffset(state, offset)
}

// Returns the occurrence at this position based on CodeMirror's built-in
// "wordAt" helper method.  This helper is a heuristic that works reasonably
// well for languages with C/Java-like identifiers, but we may want to customize
// the heurstic for other languages like Clojure where kebab-case identifiers
// are common.
function inferOccurrenceAtOffset(state: EditorState, offset: number): Occurrence | undefined {
    const identifier = state.wordAt(offset)
    // We need to ignore words that end at the requested position to match the logic
    // we use to look up occurrences in SCIP data.
    if (identifier === null || offset === identifier.to) {
        return undefined
    }
    return new Occurrence(cmSelectionToRange(state, identifier), SyntaxKind.Identifier)
}

function cmSelectionToRange(state: EditorState, selection: SelectionRange): ScipRange {
    const startLine = state.doc.lineAt(selection.from)
    const endLine = state.doc.lineAt(selection.to)
    const start = new Position(startLine.number - 1, selection.from - startLine.from)
    const end = new Position(endLine.number - 1, selection.to - endLine.from)
    return new ScipRange(start, end)
}

export function positionAtCmPosition(doc: Text, position: number): Position {
    const cmLine = doc.lineAt(position)
    const line = cmLine.number - 1
    // The lack of "- 1" at the end of the line below is intentional because it
    // makes clicking on the first character of a token have no effect.
    const character = position - cmLine.from
    return new Position(line, character)
}

export const rangeToCmSelection = (doc: Text, range: ScipRange): SelectionRange => {
    const startLine = doc.line(Math.min(doc.lines, range.start.line + 1))
    const endLine = doc.line(Math.min(doc.lines, range.end.line + 1))
    const start = startLine.from + range.start.character
    const end = Math.min(endLine.from + range.end.character, endLine.to)
    return EditorSelection.range(Math.max(0, start), Math.min(doc.length - 1, end))
}

export function contains(range: Range, position: Range['start']): boolean {
    return (
        range.start.line <= position.line &&
        range.start.character <= position.character &&
        position.line <= range.end.line &&
        position.character <= range.end.character
    )
}
