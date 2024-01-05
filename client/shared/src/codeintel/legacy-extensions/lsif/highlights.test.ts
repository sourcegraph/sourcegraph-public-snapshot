import * as assert from 'assert'

import { describe, it } from 'vitest'

import { filterLocationsForDocumentHighlights } from './highlights'
import { range1, range2, range3, range4, document } from './util.test'

describe('filterLocationsForDocumentHighlights', () => {
    it('should filter out distinct paths', () => {
        assert.deepStrictEqual(
            filterLocationsForDocumentHighlights(document, [
                { uri: new URL(document.uri), range: range1 },
                { uri: new URL(document.uri + '_distinct'), range: range2 },
                { uri: new URL(document.uri + '_distinct'), range: range3 },
                { uri: new URL(document.uri), range: range4 },
                { uri: new URL(document.uri) },
            ]),
            [{ range: range1 }, { range: range4 }]
        )
    })
})
