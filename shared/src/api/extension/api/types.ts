import * as clientType from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { Position } from '../types/position'
import { Range } from '../types/range'

/**
 * Converts from a plain object {@link clientType.Position} to an instance of {@link Position}.
 *
 * @internal
 */
export function toPosition(position: clientType.Position): Position {
    return new Position(position.line, position.character)
}

/**
 * Converts from an instance of {@link Location} to the plain object {@link clientType.Location}.
 *
 * @internal
 */
export function fromLocation(location: sourcegraph.Location): clientType.Location {
    return {
        uri: location.uri.toString(),
        range: fromRange(location.range),
    }
}

/**
 * Converts from an instance of {@link Hover} to the plain object {@link clientType.Hover}.
 *
 * @internal
 */
export function fromHover(hover: sourcegraph.Hover): clientType.Hover {
    return {
        contents: hover.contents,
        __backcompatContents: hover.__backcompatContents, // tslint:disable-line deprecation
        range: fromRange(hover.range),
    }
}

/**
 * Converts from an instance of {@link Range} to the plain object {@link clientType.Range}.
 *
 * @internal
 */
export function fromRange(range: Range | sourcegraph.Range | undefined): clientType.Range | undefined {
    if (!range) {
        return undefined
    }
    return range instanceof Range ? range.toJSON() : range
}

/**
 * Converts from an instance of {@link SearchResult} to the plain object {@link clientType.SearchResult}.
 *
 * @internal
 */
export function fromSearchResult(searchResult: sourcegraph.SearchResult): clientType.SearchResult {
    return {
        ...searchResult,
        matches: searchResult.matches.map(match => fromSearchResultMatch(match)),
    }
}

/**
 * Converts from an instance of {@link SearchResultMatch} to the plain object {@link clientType.SearchResultMatch}.
 *
 * @internal
 */
export function fromSearchResultMatch(searchMatch: sourcegraph.SearchResultMatch): clientType.SearchResultMatch {
    return {
        body: searchMatch.body,
        highlights: searchMatch.highlights.map(match => fromRange(match)) as clientType.Range[],
        url: searchMatch.url,
    }
}
