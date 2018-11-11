import * as sourcegraph from 'sourcegraph';
export declare class Location implements sourcegraph.Location {
    static isLocation(thing: any): thing is sourcegraph.Location;
    uri: sourcegraph.URI;
    range?: sourcegraph.Range;
    constructor(uri: sourcegraph.URI, rangeOrPosition?: sourcegraph.Range | sourcegraph.Position);
    toJSON(): any;
}
