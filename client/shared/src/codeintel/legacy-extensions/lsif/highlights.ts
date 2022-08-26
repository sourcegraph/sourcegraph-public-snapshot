import * as sourcegraph from '@sourcegraph/extension-api-types'

import { isDefined } from '../util/helpers'

export function filterLocationsForDocumentHighlights(
    textDocument: sourcegraph.TextDocument,
    locations: sourcegraph.Location[]
): sourcegraph.DocumentHighlight[] {
    return locations
        .filter(({ uri }) => uri.toString() === textDocument.uri)
        .map(({ range }) => range)
        .filter(isDefined)
        .map(range => ({ range }))
}
