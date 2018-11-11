import * as sourcegraph from 'sourcegraph';
import { Position } from './position';
import { Range } from './range';
export declare class Selection extends Range implements sourcegraph.Selection {
    static isSelection(thing: any): thing is Selection;
    private _anchor;
    readonly anchor: Position;
    private _active;
    readonly active: Position;
    constructor(anchor: Position, active: Position);
    constructor(anchorLine: number, anchorColumn: number, activeLine: number, activeColumn: number);
    readonly isReversed: boolean;
    toJSON(): any;
}
