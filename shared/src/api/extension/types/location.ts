import * as sourcegraph from 'sourcegraph'
import { Position } from './position'
import { Range } from './range'
import { URI } from './uri'

export class Location implements sourcegraph.Location {
    public static isLocation(thing: any): thing is sourcegraph.Location {
        if (thing instanceof Location) {
            return true
        }
        if (!thing) {
            return false
        }
        return Range.isRange((thing as Location).range) && URI.isURI((thing as Location).uri)
    }

    public readonly range?: sourcegraph.Range

    constructor(public readonly uri: sourcegraph.URI, rangeOrPosition?: sourcegraph.Range | sourcegraph.Position) {
        if (!rangeOrPosition) {
            // that's OK
        } else if (rangeOrPosition instanceof Range) {
            this.range = rangeOrPosition
        } else if (rangeOrPosition instanceof Position) {
            this.range = new Range(rangeOrPosition, rangeOrPosition)
        } else {
            throw new Error('Illegal argument')
        }
    }

    public toJSON(): any {
        return {
            uri: this.uri,
            range: this.range,
        }
    }
}
