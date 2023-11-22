import type * as sourcegraph from 'sourcegraph'

import { Position, Range } from '@sourcegraph/extension-api-classes'
import type * as clientType from '@sourcegraph/extension-api-types'

/**
 * Firefox does not like URLs being transmitted, so convert them to strings.
 *
 * https://github.com/sourcegraph/sourcegraph/issues/8928
 */
export function fromDocumentSelector(selector: sourcegraph.DocumentSelector): sourcegraph.DocumentSelector {
    return selector.map(filter =>
        typeof filter === 'string' ? filter : { ...filter, baseUri: filter.baseUri?.toString() }
    )
}

/**
 * Converts from a plain object {@link clientType.Position} to an instance of {@link Position}.
 *
 * @internal
 */
export function toPosition(position: clientType.Position): Position {
    return new Position(position.line, position.character)
}

/**
 * Converts from an instance of a badged {@link Location} to the plain object {@link clientType.Location}.
 *
 * @internal
 */
export function fromLocation(
    location: sourcegraph.Badged<sourcegraph.Location>
): sourcegraph.Badged<clientType.Location> {
    return {
        uri: location.uri.href,
        range: fromRange(location.range),
        aggregableBadges: location.aggregableBadges,
    }
}

/**
 * Converts from an instance of a badged {@link Hover} to the plain object {@link clientType.Hover}.
 *
 * @internal
 */
export function fromHover(hover: sourcegraph.Badged<sourcegraph.Hover>): sourcegraph.Badged<clientType.Hover> {
    return {
        contents: hover.contents,
        range: fromRange(hover.range),
        aggregableBadges: hover.aggregableBadges,
    }
}

/**
 * Converts from an instance of {@link Range} to the plain object {@link clientType.Range}.
 *
 * @internal
 */
function fromRange(range: Range | sourcegraph.Range | undefined): clientType.Range | undefined {
    if (!range) {
        return undefined
    }
    return range instanceof Range ? range.toJSON() : range
}
