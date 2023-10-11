import { EditorSelection, Text, type EditorState, type SelectionRange } from '@codemirror/state'

import type { Range } from '@sourcegraph/extension-api-types'
import { Occurrence, Position, Range as ScipRange, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { type HighlightIndex, syntaxHighlight } from './highlight'

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

export const isInteractiveOccurrence = (occurrence: Occurrence): boolean => {
    if (!occurrence.kind) {
        return false
    }

    return INTERACTIVE_OCCURRENCE_KINDS.has(occurrence.kind)
}

export function occurrenceAt(state: EditorState, offset: number): Occurrence | undefined {
    // First we try to get an occurrence from syntax highlighting data.
    const fromHighlighting = highlightingOccurrenceAtPosition(state, offset)
    if (fromHighlighting) {
        return fromHighlighting
    }
    // If the syntax highlighting data is incomplete then we fallback to a
    // heursitic to infer the occurrence.
    return inferOccurrenceAtPosition(state, offset)
}

// Returns the occurrence at this position based on syntax highlighting data.
// The highlighting data can come from Syntect (low-ish quality) or tree-sitter
// (better quality). When we implement semantic highlighting in the future, the
// highlighting data may come from precise indexers.
function highlightingOccurrenceAtPosition(state: EditorState, offset: number): Occurrence | undefined {
    const position = positionAtCmPosition(state.doc, offset)
    const table = state.facet(syntaxHighlight)
    for (
        let index = table.lineIndex[position.line];
        index !== undefined &&
        index < table.occurrences.length &&
        table.occurrences[index].range.start.line === position.line;
        index++
    ) {
        const occurrence = table.occurrences[index]
        if (occurrence.range.contains(position)) {
            return occurrence
        }
    }
    return undefined
}

// Returns the occurrence at this position based on CodeMirror's built-in
// "wordAt" helper method.  This helper is a heuristic that works reasonably
// well for languages with C/Java-like identifiers, but we may want to customize
// the heurstic for other languages like Clojure where kebab-case identifiers
// are common.
function inferOccurrenceAtPosition(state: EditorState, offset: number): Occurrence | undefined {
    const identifier = state.wordAt(offset)
    // We need to ignore words that end at the requested position to match the logic
    // we use to look up occurrences in SCIP data.
    if (identifier === null || offset === identifier.to) {
        return undefined
    }
    return new Occurrence(cmSelectionToRange(state, identifier), SyntaxKind.Identifier)
}

// Returns the occurrence in the provided line number that is closest to the
// provided position, compared by the character (not line). Returns undefined
// when the line has no occurrences (for example, an empty string).
export function closestOccurrenceByCharacter(
    line: number,
    table: HighlightIndex,
    position: Position,
    includeOccurrence?: (occurrence: Occurrence) => boolean
): Occurrence | undefined {
    const candidates: [Occurrence, number][] = []
    let index = table.lineIndex[line] ?? -1
    for (
        ;
        index >= 0 && index < table.occurrences.length && table.occurrences[index].range.start.line === line;
        index++
    ) {
        const occurrence = table.occurrences[index]
        if (!isInteractiveOccurrence(occurrence)) {
            continue
        }
        if (includeOccurrence && !includeOccurrence(occurrence)) {
            continue
        }
        candidates.push([occurrence, occurrence.range.characterDistance(position)])
    }
    candidates.sort(([, a], [, b]) => a - b)
    if (candidates.length > 0) {
        return candidates[0][0]
    }
    return undefined
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
