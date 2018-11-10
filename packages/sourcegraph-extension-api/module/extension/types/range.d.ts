import * as sourcegraph from 'sourcegraph';
import { Position } from './position';
export declare class Range implements sourcegraph.Range {
    static isRange(thing: any): thing is sourcegraph.Range;
    protected _start: Position;
    protected _end: Position;
    readonly start: sourcegraph.Position;
    readonly end: sourcegraph.Position;
    constructor(start: sourcegraph.Position, end: sourcegraph.Position);
    constructor(startLine: number, startColumn: number, endLine: number, endColumn: number);
    contains(positionOrRange: sourcegraph.Position | sourcegraph.Range): boolean;
    isEqual(other: sourcegraph.Range): boolean;
    intersection(other: sourcegraph.Range): sourcegraph.Range | undefined;
    union(other: sourcegraph.Range): sourcegraph.Range;
    readonly isEmpty: boolean;
    readonly isSingleLine: boolean;
    with(start?: sourcegraph.Position, end?: sourcegraph.Position): sourcegraph.Range;
    with(change: {
        start?: sourcegraph.Position;
        end?: sourcegraph.Position;
    }): sourcegraph.Range;
    toJSON(): any;
}
