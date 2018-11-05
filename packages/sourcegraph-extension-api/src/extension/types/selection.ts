import * as sourcegraph from 'sourcegraph'
import { Position } from './position'
import { Range } from './range'

export class Selection extends Range implements sourcegraph.Selection {
    public static isSelection(thing: any): thing is Selection {
        if (thing instanceof Selection) {
            return true
        }
        if (!thing) {
            return false
        }
        return (
            Range.isRange(thing) &&
            Position.isPosition((thing as Selection).anchor) &&
            Position.isPosition((thing as Selection).active) &&
            typeof (thing as Selection).isReversed === 'boolean'
        )
    }

    private _anchor: Position

    public get anchor(): Position {
        return this._anchor
    }

    private _active: Position

    public get active(): Position {
        return this._active
    }

    constructor(anchor: Position, active: Position)
    constructor(anchorLine: number, anchorColumn: number, activeLine: number, activeColumn: number)
    constructor(
        anchorLineOrAnchor: number | Position,
        anchorColumnOrActive: number | Position,
        activeLine?: number,
        activeColumn?: number
    ) {
        let anchor: Position | undefined
        let active: Position | undefined

        if (
            typeof anchorLineOrAnchor === 'number' &&
            typeof anchorColumnOrActive === 'number' &&
            typeof activeLine === 'number' &&
            typeof activeColumn === 'number'
        ) {
            anchor = new Position(anchorLineOrAnchor, anchorColumnOrActive)
            active = new Position(activeLine, activeColumn)
        } else if (anchorLineOrAnchor instanceof Position && anchorColumnOrActive instanceof Position) {
            anchor = anchorLineOrAnchor
            active = anchorColumnOrActive
        }

        if (!anchor || !active) {
            throw new Error('Invalid arguments')
        }

        super(anchor, active)

        this._anchor = anchor
        this._active = active
    }

    public get isReversed(): boolean {
        return this._anchor === this._end
    }

    public toJSON(): any {
        return {
            start: this.start,
            end: this.end,
            active: this.active,
            anchor: this.anchor,
        }
    }
}
