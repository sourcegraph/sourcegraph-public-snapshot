// Restates vscode.Range
interface TrackedRange<
    Self extends TrackedRange<Self, PositionType>,
    PositionType extends TrackedPosition<PositionType>
> {
    start: PositionType
    end: PositionType
    with(start: PositionType, end: PositionType): Self
}

// Restates vscode.Position
interface TrackedPosition<Self extends TrackedPosition<Self>> {
    character: number
    line: number
    isAfter(other: Self): boolean
    isAfterOrEqual(other: Self): boolean
    isBefore(other: Self): boolean
    isBeforeOrEqual(other: Self): boolean
    isEqual(other: Self): boolean
    translate(lineDelta?: number, characterDelta?: number): Self
}

// Restates vscode.TextDocumentContentChangeEvent
interface TextChange<
    RangeType extends TrackedRange<RangeType, PositionType>,
    PositionType extends TrackedPosition<PositionType>
> {
    range: RangeType
    text: string
}

export function updateRange<
    RangeType extends TrackedRange<RangeType, PositionType>,
    PositionType extends TrackedPosition<PositionType>
>(range: RangeType, change: TextChange<RangeType, PositionType>): RangeType | null {
    const lines = change.text.split(/\r\n|\r|\n/m)
    const insertedLastLine = lines.at(-1)?.length
    if (typeof insertedLastLine === 'undefined') {
        throw new TypeError('unreachable') // Any string .split produces a non-empty array.
    }
    const insertedLineBreaks = lines.length - 1

    // Handle character deletions
    // ...after
    if (change.range.start.isAfterOrEqual(range.end)) {
        return range
    }
    // ...before
    // TODO: Insertion *at* the start, including a newline, eats the line by change.range.start.character
    if (change.range.end.isBeforeOrEqual(range.start)) {
        range = range.with(
            range.start.translate(
                change.range.start.line - change.range.end.line + insertedLineBreaks,
                change.range.end.line === range.start.line
                    ? insertedLastLine +
                          -change.range.end.character +
                          (insertedLineBreaks === 0 ? change.range.start.character : 0)
                    : 0
            ),
            range.end.translate(
                change.range.start.line - change.range.end.line + insertedLineBreaks,
                change.range.end.line === range.end.line
                    ? insertedLastLine -
                          change.range.end.character +
                          (insertedLineBreaks === 0 ? change.range.start.character : 0)
                    : 0
            )
        )
    }
    // ...around
    else if (change.range.start.isBeforeOrEqual(range.start) && change.range.end.isAfterOrEqual(range.end)) {
        return null
    }
    // ...within
    else if (change.range.start.isAfterOrEqual(range.start) && change.range.end.isBeforeOrEqual(range.end)) {
        range = range.with(
            range.start,
            range.end.translate(
                change.range.start.line - change.range.end.line + insertedLineBreaks,
                change.range.end.line === range.end.line
                    ? change.range.start.character - change.range.end.character + insertedLastLine
                    : 0
            )
        )
    }
    // ...overlapping start
    else if (change.range.end.isBefore(range.end)) {
        range = range.with(
            // Move the start of the decoration to the end of the change
            change.range.end.translate(
                change.range.start.line - change.range.end.line,
                change.range.start.character - change.range.end.character + insertedLastLine
            ),
            // Adjust the end of the decoration for the range deletion
            range.end.translate(
                change.range.start.line - change.range.end.line,
                change.range.end.line === range.end.line
                    ? change.range.start.character - change.range.end.character + insertedLastLine
                    : 0
            )
        )
    }
    // ...overlapping end
    else {
        range = range.with(
            range.start,
            // Move the end of the decoration to the start of the change
            change.range.start
        )
    }
    return range.start.isEqual(range.end) ? null : range
}
