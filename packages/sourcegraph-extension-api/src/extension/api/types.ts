import * as sourcegraph from 'sourcegraph'
import * as plain from '../../protocol/plainTypes'
import { Position } from '../types/position'
import { Range } from '../types/range'

/**
 * Converts from a plain object {@link plain.Position} to an instance of {@link Position}.
 *
 * @internal
 */
export function toPosition(position: plain.Position): Position {
    return new Position(position.line, position.character)
}

/**
 * Converts from an instance of {@link Location} to the plain object {@link plain.Location}.
 *
 * @internal
 */
export function fromLocation(location: sourcegraph.Location): plain.Location {
    return {
        uri: location.uri.toString(),
        range: fromRange(location.range),
    }
}

/**
 * Converts from an instance of {@link Hover} to the plain object {@link plain.Hover}.
 *
 * @internal
 */
export function fromHover(hover: sourcegraph.Hover): plain.Hover {
    return {
        contents: hover.contents,
        __backcompatContents: hover.__backcompatContents, // tslint:disable-line deprecation
        range: fromRange(hover.range),
    }
}

/**
 * Converts from an instance of {@link Range} to the plain object {@link plain.Range}.
 *
 * @internal
 */
export function fromRange(range: Range | sourcegraph.Range | undefined): plain.Range | undefined {
    if (!range) {
        return undefined
    }
    return range instanceof Range ? range.toJSON() : range
}
