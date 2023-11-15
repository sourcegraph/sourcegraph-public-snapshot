import type { Line, SelectionRange, Text } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'

import type { LineOrPositionOrRange } from '@sourcegraph/common'
import type { Position } from '@sourcegraph/extension-api-types'
import { Range } from '@sourcegraph/shared/src/codeintel/scip'
import {
    type BlobViewState,
    type UIPositionSpec,
    type UIRangeSpec,
    toPrettyBlobURL,
} from '@sourcegraph/shared/src/util/url'

import type { Location } from './codeintel/api'

/**
 * The MouseEvent uses numbers to indicate which button was pressed.
 * See https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/button#value
 */
export const MOUSE_MAIN_BUTTON = 0

export function zeroToOneBasedPosition(position: Position): { line: number; character: number } {
    return {
        line: position.line + 1,
        character: position.character + 1,
    }
}

export function zeroToOneBasedRange(range: { start: Position; end: Position }): {
    start: { line: number; character: number }
    end: { line: number; character: number }
} {
    return {
        start: zeroToOneBasedPosition(range.start),
        end: zeroToOneBasedPosition(range.end),
    }
}

/**
 * Converts 0-based line/character positions to document offsets.
 * Returns null if the position cannot be mapped to a valid offset within the
 * document.
 */
export function positionToOffset(textDocument: Text, position: Position): number | null {
    // Position is 0-based
    const lineNumber = position.line + 1
    if (lineNumber > textDocument.lines) {
        return null
    }
    const line = textDocument.line(lineNumber)
    const offset = line.from + position.character

    return offset <= line.to ? offset : null
}

/*
 * Converts 1-based line/character positions to document offsets.
 * Returns null if the position cannot be mapped to a valid offset within the
 * document.
 */
export function uiPositionToOffset(
    textDocument: Text,
    position: UIPositionSpec['position'],
    line?: Line
): number | null {
    const lineNumber = position.line
    if (lineNumber > textDocument.lines) {
        return null
    }
    if (!line) {
        line = textDocument.line(position.line)
    }
    const offset = line.from + position.character - 1

    return offset <= line.to ? offset : null
}

/*
 * Converts 1-based line/character positions to document ranges.
 * Returns null if the position cannot be mapped to a valid offset within the
 * document.
 */
export function lprToRange(textDocument: Text, position: LineOrPositionOrRange): { from: number; to: number } | null {
    const { line, character, endLine, endCharacter } = position
    if (!line || line > textDocument.lines) {
        return null
    }

    const fromLine = textDocument.line(line)
    let from = fromLine.from

    if (character) {
        from += character - 1
    }

    let to = from

    if (endLine && endLine < textDocument.lines) {
        to = textDocument.line(endLine).from
    }

    if (endCharacter) {
        to += endCharacter - 1
    }

    return { from, to }
}

/**
 * Converts document offsets to 1-based line/character positions.
 */
export function offsetToUIPosition(textDocument: Text, from: number): UIPositionSpec['position']
export function offsetToUIPosition(textDocument: Text, from: number, to: number): UIRangeSpec['range']
export function offsetToUIPosition(
    textDocument: Text,
    from: number,
    to?: number
): UIRangeSpec['range'] | UIPositionSpec['position'] {
    const startLine = textDocument.lineAt(from)
    const startCharacter = from - startLine.from + 1

    const startPosition: Position = {
        line: startLine.number,
        character: startCharacter,
    }

    if (to !== undefined) {
        const endLine = to <= startLine.to ? startLine : textDocument.lineAt(to)
        const endCharacter = to - endLine.from + 1

        return {
            start: startPosition,
            end: {
                line: endLine.number,
                character: endCharacter,
            },
        }
    }

    return startPosition
}

export function sortRangeValuesByStart<T extends { range: { start: Position } }>(values: T[]): T[] {
    return values.sort(({ range: rangeA }, { range: rangeB }) =>
        rangeA.start.line === rangeB.start.line
            ? rangeA.start.character - rangeB.start.character
            : rangeA.start.line - rangeB.start.line
    )
}

/**
 * Returns the document offset at the provided coordindates or null if there is
 * no text underneath these coordinates.
 *
 * It seems when using `posAtCoords` CodeMirror returns the document position
 * _closest_ to the coordinates. This has the unfortunate effect that hovering
 * over empty parts a line will find the position of the closest character next
 * to it, which we do not want.
 * To ensure that we only consider positions of actual words/characters we
 * perform the inverse conversion and compare the results.
 * This is also done by CodeMirror's own hover tooltip plugin.
 */
export function preciseOffsetAtCoords(view: EditorView, coords: { x: number; y: number }): number | null {
    const offset = view.posAtCoords(coords)
    if (offset === null) {
        return null
    }
    const offsetCords = view.coordsAtPos(offset)
    if (
        offsetCords === null ||
        coords.y < offsetCords.top ||
        coords.y > offsetCords.bottom ||
        coords.x < offsetCords.left - view.defaultCharacterWidth ||
        coords.x > offsetCords.right + view.defaultCharacterWidth
    ) {
        return null
    }
    return offset
}

/**
 * Gets the word at the precise offsets at the provided coordinates.
 * See {@link preciseOffsetAtCoords} for more information.
 */
export function preciseWordAtCoords(
    view: EditorView,
    coords: { x: number; y: number }
): { from: number; to: number } | null {
    const offset = preciseOffsetAtCoords(view, coords)
    if (offset !== null) {
        return view.state.wordAt(offset)
    }

    return null
}

export function isSelectionInsideDocument(selection: SelectionRange, doc: Text): boolean {
    return selection.from >= 0 && selection.to < doc.length
}

/**
 * Verifies that the provided 1-based range is within the document range.
 */
export function isValidLineRange(
    range: { line: number; character?: number; endLine?: number; endCharacter?: number },
    textDocument: Text
): boolean {
    const { lines } = textDocument

    // Return early if the document doesn't have as many lines
    if ((range.endLine ?? range.line) > lines) {
        return false
    }

    {
        // Some juggling to make Typescript happy (passing range directly
        // doesn't work)
        const { character, line } = range
        if (character && uiPositionToOffset(textDocument, { line, character }) === null) {
            return false
        }
    }

    if (range.endLine && range.endCharacter) {
        return uiPositionToOffset(textDocument, { line: range.endLine, character: range.endCharacter }) !== null
    }

    return true
}

export function locationToURL(
    documentInfo: { repoName: string; filePath: string; commitID: string; revision?: string },
    location: Location,
    viewState?: BlobViewState
): string {
    const { range, filePath, repoName, revision: locationRevision } = location

    // Try to preserve the non-canonical revision (branch name or empty revision)
    // when possible.
    const preserveNonCanonicalRevision =
        documentInfo.repoName === repoName && // Destination location is within the same repo.
        documentInfo.revision !== documentInfo.commitID && // Current URL is non-canonical/non-commit revision.
        documentInfo.commitID === locationRevision // Destination revision is the as as current revision.
    const revision = preserveNonCanonicalRevision ? documentInfo.revision : locationRevision

    return toPrettyBlobURL({
        repoName,
        revision,
        filePath,
        position: { line: range.start.line + 1, character: range.start.character + 1 },
        range: location.range ? Range.fromExtensions(location.range).withIncrementedValues() : undefined,
        viewState,
    })
}

// Returns true if this event is "regular", meaning the user is not holding down
// modifier keys or clicking with a non-main button.
export function isRegularEvent(event: MouseEvent | KeyboardEvent): boolean {
    return (
        ('button' in event ? event.button === MOUSE_MAIN_BUTTON : true) &&
        !event.metaKey &&
        !event.shiftKey &&
        !event.ctrlKey
    )
}

export function browserIsFirefox(): boolean {
    return navigator.userAgent.includes('Firefox')
}
