import * as sourcegraph from 'sourcegraph';
export declare class Position {
    static min(position: sourcegraph.Position, ...positions: sourcegraph.Position[]): Position;
    static min(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined;
    static max(position: sourcegraph.Position, ...positions: sourcegraph.Position[]): Position;
    static max(...positions: sourcegraph.Position[]): sourcegraph.Position | undefined;
    static isPosition(other: any): other is sourcegraph.Position;
    private _line;
    private _character;
    readonly line: number;
    readonly character: number;
    constructor(line: number, character: number);
    isBefore(other: sourcegraph.Position): boolean;
    isBeforeOrEqual(other: sourcegraph.Position): boolean;
    isAfter(other: sourcegraph.Position): boolean;
    isAfterOrEqual(other: sourcegraph.Position): boolean;
    isEqual(other: sourcegraph.Position): boolean;
    compareTo(other: sourcegraph.Position): number;
    translate(lineDelta?: number, characterDelta?: number): sourcegraph.Position;
    translate(change: {
        lineDelta?: number;
        characterDelta?: number;
    }): sourcegraph.Position;
    with(line?: number, character?: number): sourcegraph.Position;
    with(change: {
        line?: number;
        character?: number;
    }): sourcegraph.Position;
    toJSON(): any;
}
