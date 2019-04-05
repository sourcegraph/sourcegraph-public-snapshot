import * as clientType from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { illegalArgument } from './errors'
import { Position } from './position'

export class Range implements sourcegraph.Range {
    public static isRange(thing: any): thing is sourcegraph.Range {
        if (thing instanceof Range) {
            return true
        }
        if (!thing) {
            return false
        }
        return Position.isPosition((thing as Range).start) && Position.isPosition(thing.end as Range)
    }

    protected _start: Position
    protected _end: Position

    public get start(): sourcegraph.Position {
        return this._start
    }

    public get end(): sourcegraph.Position {
        return this._end
    }

    constructor(start: sourcegraph.Position, end: sourcegraph.Position)
    constructor(startLine: number, startColumn: number, endLine: number, endColumn: number)
    constructor(
        startLineOrStart: number | sourcegraph.Position,
        startColumnOrEnd: number | sourcegraph.Position,
        endLine?: number,
        endColumn?: number
    ) {
        let start: Position | undefined
        let end: Position | undefined

        if (
            typeof startLineOrStart === 'number' &&
            typeof startColumnOrEnd === 'number' &&
            typeof endLine === 'number' &&
            typeof endColumn === 'number'
        ) {
            start = new Position(startLineOrStart, startColumnOrEnd)
            end = new Position(endLine, endColumn)
        } else if (startLineOrStart instanceof Position && startColumnOrEnd instanceof Position) {
            start = startLineOrStart
            end = startColumnOrEnd
        }

        if (!start || !end) {
            throw new Error('Invalid arguments')
        }

        if (start.isBefore(end)) {
            this._start = start
            this._end = end
        } else {
            this._start = end
            this._end = start
        }
    }

    public contains(positionOrRange: sourcegraph.Position | sourcegraph.Range): boolean {
        if (positionOrRange instanceof Range) {
            return this.contains(positionOrRange._start) && this.contains(positionOrRange._end)
        }
        if (positionOrRange instanceof Position) {
            if (positionOrRange.isBefore(this._start)) {
                return false
            }
            if (this._end.isBefore(positionOrRange)) {
                return false
            }
            return true
        }
        return false
    }

    public isEqual(other: sourcegraph.Range): boolean {
        return this._start.isEqual(other.start) && this._end.isEqual(other.end)
    }

    public intersection(other: sourcegraph.Range): sourcegraph.Range | undefined {
        const start = Position.max(other.start, this._start)
        const end = Position.min(other.end, this._end)
        if (start.isAfter(end)) {
            // this happens when there is no overlap:
            // |-----|
            //          |----|
            return undefined
        }
        return new Range(start, end)
    }

    public union(other: sourcegraph.Range): sourcegraph.Range {
        if (this.contains(other)) {
            return this
        }
        if (other.contains(this)) {
            return other
        }
        const start = Position.min(other.start, this._start)
        const end = Position.max(other.end, this.end)
        return new Range(start, end)
    }

    public get isEmpty(): boolean {
        return this._start.isEqual(this._end)
    }

    public get isSingleLine(): boolean {
        return this._start.line === this._end.line
    }

    public with(start?: sourcegraph.Position, end?: sourcegraph.Position): sourcegraph.Range
    public with(change: { start?: sourcegraph.Position; end?: sourcegraph.Position }): sourcegraph.Range
    public with(
        startOrChange: sourcegraph.Position | undefined | { start?: sourcegraph.Position; end?: sourcegraph.Position },
        end: sourcegraph.Position = this.end
    ): sourcegraph.Range {
        if (startOrChange === null || end === null) {
            throw illegalArgument()
        }

        let start: sourcegraph.Position
        if (!startOrChange) {
            start = this.start
        } else if (Position.isPosition(startOrChange)) {
            start = startOrChange
        } else {
            start = startOrChange.start || this.start
            end = startOrChange.end || this.end
        }

        if (start.isEqual(this._start) && end.isEqual(this.end)) {
            return this
        }
        return new Range(start, end)
    }

    public toJSON(): any {
        return { start: this._start.toJSON(), end: this._end.toJSON() }
    }

    public toPlain(): clientType.Range {
        return {
            start: { line: this._start.line, character: this._start.character },
            end: { line: this._end.line, character: this._end.character },
        }
    }
}
