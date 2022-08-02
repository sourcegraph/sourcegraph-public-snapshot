import { Text } from '@codemirror/state'

import { Position, Range } from '@sourcegraph/extension-api-types'
import { UIPositionSpec, UIRangeSpec } from '@sourcegraph/shared/src/util/url'

/**
 * Converts 0-based line/character positions to document offsets.
 */
export function positionToOffset(textDocument: Text, position: Position): number {
    // Position is 0-based
    return textDocument.line(position.line + 1).from + position.character
}

/*
 * Converts 1-based line/character positions to document offsets.
 */
export function uiPositionToOffset(
    textDocument: Text,
    position: UIPositionSpec['position'],
    line = textDocument.line(position.line)
): number {
    return line.from + position.character - 1
}

/**
 * Converts document offsets 1-based line/character positions.
 */
export function offsetToPosition(textDocument: Text, from: number): UIPositionSpec['position']
export function offsetToPosition(textDocument: Text, from: number, to: number): UIRangeSpec['range']
export function offsetToPosition(
    textDocument: Text,
    from: number,
    to?: number
): UIRangeSpec['range'] | UIPositionSpec['position'] {
    const startLine = textDocument.lineAt(from)
    const startCharacter = Math.max(0, from - startLine.from) + 1

    const startPosition: Position = {
        line: startLine.number,
        character: startCharacter,
    }

    if (to !== undefined) {
        const endLine = to <= startLine.to ? startLine : textDocument.lineAt(to)
        const endCharacter = Math.max(0, to - endLine.to) + 1

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

export function viewPortChanged(
    previous: { from: number; to: number } | null,
    next: { from: number; to: number }
): boolean {
    return previous?.from !== next.from || previous.to !== next.to
}

export function sortRangeValuesByStart<T extends { range: Range }>(values: T[]): T[] {
    return values.sort(({ range: rangeA }, { range: rangeB }) =>
        rangeA.start.line === rangeB.start.line
            ? rangeA.start.character - rangeB.start.character
            : rangeA.start.line - rangeB.start.line
    )
}
