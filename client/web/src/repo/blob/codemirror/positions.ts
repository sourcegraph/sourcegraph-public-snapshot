import { EditorSelection, EditorState, SelectionRange } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { Occurrence, Position, Range } from '@sourcegraph/shared/src/codeintel/scip'

import { HighlightIndex, syntaxHighlight } from './highlight'
import { isInteractiveOccurrence } from './tokens-as-links'

export function occurrenceAtPosition(
    state: EditorState,
    position: Position
): { occurrence: Occurrence; position: Position } | undefined {
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
            return { occurrence, position }
        }
    }
    return
}

export function closestOccurrence(
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

export function occurrenceAtEvent(
    view: EditorView,
    event: MouseEvent
): { occurrence: Occurrence; position: Position } | undefined {
    const atEvent = positionAtEvent(view, event)
    if (!atEvent) {
        return
    }
    const { position } = atEvent
    const occurrence = occurrenceAtPosition(view.state, position)
    if (!occurrence) {
        return
    }
    return { ...occurrence }
}

export function positionAtEvent(view: EditorView, event: MouseEvent): { position: Position } | undefined {
    const position = view.posAtCoords({ x: event.clientX, y: event.clientY })
    if (position === null) {
        return
    }
    event.preventDefault()
    return { position: positionAtCmPosition(view, position) }
}

export function positionAtCmPosition(view: EditorView, position: number): Position {
    const cmLine = view.state.doc.lineAt(position)
    const line = cmLine.number - 1
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

export const rangeToSelection = (state: EditorState, range: Range): SelectionRange => {
    const startLine = state.doc.line(range.start.line + 1)
    const endLine = state.doc.line(range.end.line + 1)
    const start = startLine.from + range.start.character
    const end = Math.min(endLine.from + range.end.character, endLine.to)
    return EditorSelection.range(start, end)
}
