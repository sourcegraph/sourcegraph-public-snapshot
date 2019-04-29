import * as sourcegraph from 'sourcegraph'
import { illegalArgument } from './errors'

export class Position implements sourcegraph.Position {
    public static min(position: sourcegraph.Position, ...positions: sourcegraph.Position[]): Position
    public static min(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined
    public static min(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined {
        let result = positions.pop()
        if (result === undefined) {
            return undefined
        }
        for (const p of positions) {
            if (p.isBefore(result)) {
                result = p
            }
        }
        return result
    }

    public static max(position: sourcegraph.Position, ...positions: sourcegraph.Position[]): Position
    public static max(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined
    public static max(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined {
        let result = positions.pop()
        if (result === undefined) {
            return undefined
        }
        for (const p of positions) {
            if (p.isAfter(result)) {
                result = p
            }
        }
        return result
    }

    public static isPosition(other: any): other is sourcegraph.Position {
        if (!other) {
            return false
        }
        if (other instanceof Position) {
            return true
        }
        const { line, character } = other as sourcegraph.Position
        if (typeof line === 'number' && typeof character === 'number') {
            return true
        }
        return false
    }

    private _line: number
    private _character: number

    public get line(): number {
        return this._line
    }

    public get character(): number {
        return this._character
    }

    constructor(line: number, character: number) {
        if (line < 0) {
            throw illegalArgument('line must be non-negative')
        }
        if (character < 0) {
            throw illegalArgument('character must be non-negative')
        }
        this._line = line
        this._character = character
    }

    public isBefore(other: sourcegraph.Position): boolean {
        if (this._line < other.line) {
            return true
        }
        if (other.line < this._line) {
            return false
        }
        return this._character < other.character
    }

    public isBeforeOrEqual(other: sourcegraph.Position): boolean {
        if (this._line < other.line) {
            return true
        }
        if (other.line < this._line) {
            return false
        }
        return this._character <= other.character
    }

    public isAfter(other: sourcegraph.Position): boolean {
        return !this.isBeforeOrEqual(other)
    }

    public isAfterOrEqual(other: sourcegraph.Position): boolean {
        return !this.isBefore(other)
    }

    public isEqual(other: sourcegraph.Position): boolean {
        return this._line === other.line && this._character === other.character
    }

    public compareTo(other: sourcegraph.Position): number {
        if (this._line < other.line) {
            return -1
        }
        if (this._line > other.line) {
            return 1
        }
        // equal line
        if (this._character < other.character) {
            return -1
        }
        if (this._character > other.character) {
            return 1
        }
        // equal line and character
        return 0
    }

    public translate(lineDelta?: number, characterDelta?: number): sourcegraph.Position
    public translate(change: { lineDelta?: number; characterDelta?: number }): sourcegraph.Position
    public translate(
        lineDeltaOrChange: number | undefined | { lineDelta?: number; characterDelta?: number },
        characterDelta = 0
    ): Position {
        if (lineDeltaOrChange === null || characterDelta === null) {
            throw illegalArgument()
        }

        let lineDelta: number
        if (typeof lineDeltaOrChange === 'undefined') {
            lineDelta = 0
        } else if (typeof lineDeltaOrChange === 'number') {
            lineDelta = lineDeltaOrChange
        } else {
            lineDelta = typeof lineDeltaOrChange.lineDelta === 'number' ? lineDeltaOrChange.lineDelta : 0
            characterDelta = typeof lineDeltaOrChange.characterDelta === 'number' ? lineDeltaOrChange.characterDelta : 0
        }

        if (lineDelta === 0 && characterDelta === 0) {
            return this
        }
        return new Position(this.line + lineDelta, this.character + characterDelta)
    }

    public with(line?: number, character?: number): sourcegraph.Position
    public with(change: { line?: number; character?: number }): sourcegraph.Position
    public with(
        lineOrChange: number | undefined | { line?: number; character?: number },
        character: number = this.character
    ): sourcegraph.Position {
        if (lineOrChange === null || character === null) {
            throw illegalArgument()
        }

        let line: number
        if (typeof lineOrChange === 'undefined') {
            line = this.line
        } else if (typeof lineOrChange === 'number') {
            line = lineOrChange
        } else {
            line = typeof lineOrChange.line === 'number' ? lineOrChange.line : this.line
            character = typeof lineOrChange.character === 'number' ? lineOrChange.character : this.character
        }

        if (line === this.line && character === this.character) {
            return this
        }
        return new Position(line, character)
    }

    public toJSON(): any {
        return { line: this.line, character: this.character }
    }
}
