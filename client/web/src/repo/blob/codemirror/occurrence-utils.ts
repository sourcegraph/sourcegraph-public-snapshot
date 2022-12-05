import { EditorSelection, EditorState, SelectionRange } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { Occurrence, Position, Range, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { HighlightIndex, syntaxHighlight } from './highlight'
import { fallbackOccurrences } from './token-selection/selections'
import { preciseOffsetAtCoords } from './utils'

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

export function occurrenceAtMouseEvent(
    view: EditorView,
    event: MouseEvent
): { occurrence: Occurrence; position: Position } | undefined {
    const position = positionAtMouseEvent(view, event)
    if (!position) {
        return
    }
    const occurrence = occurrenceAtPosition(view.state, position)
    if (!occurrence) {
        return
    }
    return { occurrence, position }
}

export function positionAtMouseEvent(view: EditorView, event: MouseEvent): Position | undefined {
    const position = preciseOffsetAtCoords(view, { x: event.clientX, y: event.clientY })
    if (position === null) {
        return
    }
    return positionAtCmPosition(view, position)
}

export function occurrenceAtPosition(state: EditorState, position: Position): Occurrence | undefined {
    // First we try to get an occurrence from syntax highlighting data.
    const fromHighlighting = highlightingOccurrenceAtPosition(state, position)
    if (fromHighlighting) {
        return fromHighlighting
    }
    // If the syntax highlighting data is incomplete then we fallback to a
    // heursitic to infer the occurrence.
    return inferOccurrenceAtPosition(state, position)
}

// Returns the occurrence at this position based on syntax highlighting data.
// The highlighting data can come from Syntect (low-ish quality) or tree-sitter
// (better quality). When we implement semantic highlighting in the future, the
// highlighting data may come from precise indexers.
export function highlightingOccurrenceAtPosition(state: EditorState, position: Position): Occurrence | undefined {
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
export function inferOccurrenceAtPosition(state: EditorState, position: Position): Occurrence | undefined {
    const fallback = state.field(fallbackOccurrences)
    const cmLine = state.doc.line(position.line + 1)
    const cmPosition = cmLine.from + position.character + 1
    // We rely on `Occurrence` reference equality in some downstream facets so
    // it's important to reuse instances between invocations.
    const fromCache = fallback.get(cmPosition)
    if (fromCache) {
        return fromCache
    }
    const identifier = state.wordAt(cmPosition)
    if (identifier === null) {
        return undefined
    }
    const freshOccurrence = new Occurrence(cmSelectionToRange(state, identifier), SyntaxKind.Identifier)
    for (let index = identifier.from; index < identifier.to; index++) {
        // The cache is keyed by CodeMirror numeric positions to enable cheapa and simple lookups.
        fallback.set(index, freshOccurrence)
    }
    return freshOccurrence
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

export function positionAtCmPosition(view: EditorView, position: number): Position {
    const cmLine = view.state.doc.lineAt(position)
    const line = cmLine.number - 1
    // The lack of "- 1" at the end of the line below is intentional because it
    // makes clicking on the first charcter of a token have no effect.
    const character = position - cmLine.from
    return new Position(line, character)
}

export function cmSelectionToRange(state: EditorState, selection: SelectionRange): Range {
    const startLine = state.doc.lineAt(selection.from)
    const endLine = state.doc.lineAt(selection.to)
    const start = new Position(startLine.number - 1, selection.from - startLine.from)
    const end = new Position(endLine.number - 1, selection.to - endLine.from)
    return new Range(start, end)
}

export const rangeToCmSelection = (state: EditorState, range: Range): SelectionRange => {
    const startLine = state.doc.line(Math.min(state.doc.lines, range.start.line + 1))
    const endLine = state.doc.line(Math.min(state.doc.lines, range.end.line + 1))
    const start = startLine.from + range.start.character
    const end = Math.min(endLine.from + range.end.character, endLine.to)
    return EditorSelection.range(Math.max(0, start), Math.min(state.doc.length - 1, end))
}

// Wrapper arounds a `Map<Occurrence, T>` with special support to insert undefined values.
export class OccurrenceMap<T> {
    constructor(private readonly occurrences: Map<Occurrence, T>, private readonly emptyValue: T) {}
    public withOccurrence(occurrence: Occurrence, value?: T): OccurrenceMap<T> {
        this.occurrences.set(occurrence, value ?? this.emptyValue)
        return new OccurrenceMap(this.occurrences, this.emptyValue)
    }
    public get(occurrence: Occurrence | null): { value?: T; hasOccurrence: boolean } {
        if (occurrence === null) {
            return { hasOccurrence: false }
        }
        const value = this.occurrences.get(occurrence)
        if (value === undefined) {
            return { hasOccurrence: false }
        }
        if (value === this.emptyValue) {
            return { hasOccurrence: true }
        }
        return { value, hasOccurrence: true }
    }
}
