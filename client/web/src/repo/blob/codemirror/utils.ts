import { Text } from '@codemirror/state'

import { Position, Range } from '@sourcegraph/extension-api-types'

/**
 * Converts line/character positions to document offsets.
 */
export function positionToOffset(textDocument: Text, position: Position): number {
    // Position is 0-based
    return textDocument.line(position.line + 1).from + position.character
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
