/* ------------------------------------------------------------------------------------------------
 * A collection of mock classes for testings that require the vscode package.
 *-----------------------------------------------------------------------------------------------*/

// Mock vscode.Position
export class Position {
    constructor(public line: number, public character: number) {}

    public isBefore(position: Position): boolean {
        return this.line < position.line || (this.line === position.line && this.character < position.character)
    }

    public isEqual(position: Position): boolean {
        return this.line === position.line && this.character === position.character
    }

    public isBeforeOrEqual(position: Position): boolean {
        return this.isBefore(position) || this.isEqual(position)
    }

    public isAfter(position: Position): boolean {
        return !this.isBeforeOrEqual(position)
    }

    public isAfterOrEqual(position: Position): boolean {
        return !this.isBefore(position)
    }

    public compareTo(other: Position): number {
        return this.line === other.line ? this.character - other.character : this.line - other.line
    }

    public translate(lineDelta?: number, characterDelta?: number): Position {
        return new Position(this.line + (lineDelta || 0), this.character + (characterDelta || 0))
    }

    public translateChange(change: { lineDelta?: number; characterDelta?: number }): Position {
        return new Position(this.line + (change.lineDelta || 0), this.character + (change.characterDelta || 0))
    }

    public with(line?: number, character?: number): Position {
        return new Position(line || this.line, character || this.character)
    }

    public withChange(change: { line?: number; character?: number }): Position {
        return new Position(change.line || this.line, change.character || this.character)
    }
}

// Mock vscode.Range
export class Range {
    public readonly start: Position
    public readonly end: Position
    public isEmpty: boolean
    public isSingleLine: boolean

    constructor(
        public startLine: number,
        public startCharacter: number,
        public endLine: number,
        public endCharacter: number
    ) {
        this.start = new Position(startLine, startCharacter)
        this.end = new Position(endLine, endCharacter)
        this.isEmpty = startLine === endLine && startCharacter === endCharacter
        this.isSingleLine = startLine === endLine
    }

    public contains(position: Position): boolean {
        return (
            this.start.line <= position.line &&
            this.start.character <= position.character &&
            this.end.line >= position.line &&
            this.end.character >= position.character
        )
    }

    public isEqual(range: Range): boolean {
        return (
            this.start.line === range.start.line &&
            this.start.character === range.start.character &&
            this.end.line === range.end.line &&
            this.end.character === range.end.character
        )
    }

    public isBefore(range: Range): boolean {
        return (
            this.end.line < range.start.line ||
            (this.end.line === range.start.line && this.end.character <= range.start.character)
        )
    }

    public intersection(range: Range): Range | null {
        if (this.isBefore(range) || range.isBefore(this)) {
            return null
        }
        const start = this.start.isBefore(range.start) ? range.start : this.start
        const end = this.end.isBefore(range.end) ? this.end : range.end
        return new Range(start.line, start.character, end.line, end.character)
    }

    public union(range: Range): Range {
        const start = this.start.isBefore(range.start) ? this.start : range.start
        const end = this.end.isBefore(range.end) ? range.end : this.end
        return new Range(start.line, start.character, end.line, end.character)
    }

    public with(start?: Position, end?: Position): Range {
        if (start === undefined) {
            start = this.start
        }
        if (end === undefined) {
            end = this.end
        }
        if (start === this.start && end === this.end) {
            return this
        }
        return new Range(start.line, start.character, end.line, end.character)
    }

    public withChange(change: { start: Position; end: Position }): Range {
        const start = change.start || this.start
        const end = change.end || this.end
        return new Range(start.line, start.character, end.line, end.character)
    }
}
