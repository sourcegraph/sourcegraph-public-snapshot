import * as assert from 'assert'

import { describe, it } from '@jest/globals'

import type * as sourcegraph from '../api'

import { searchStencil } from './providers'
import { document, position, range1, range2, range3, range4, range5, range6 } from './util.test'

describe('stencil', () => {
    it('should find ranges in stencils', async () => {
        const run = (ranges: sourcegraph.Range[] | undefined) =>
            searchStencil(document.uri, position, () => Promise.resolve(ranges))

        assert.deepEqual(await run(undefined), 'unknown')
        assert.deepEqual(await run([]), 'miss')
        assert.deepEqual(await run([range1, range2, range3]), 'miss')
        assert.deepEqual(await run([range1, range2, range3, range6]), 'miss')
        assert.deepEqual(await run([range1, range2, range3, range4]), 'hit')
        assert.deepEqual(await run([range5]), 'hit')
    })
})
