import * as assert from 'assert'

import * as scip from '../../scip'
import * as sourcegraph from '../api'

import { nodeToLocation } from './locations'

describe('nodeToLocation', () => {
    it('converts to a location', () => {
        const range = scip.Range.fromNumbers(10, 12, 10, 15)

        const location = nodeToLocation(
            {
                uri: 'git://github.com/baz/bonk?c7ad68d72ef7b0d5aac07c22e86fef05d38b06da#source.ts',
            } as sourcegraph.TextDocument,
            {
                resource: {
                    repository: { name: 'github.com/foo/bar' },
                    commit: { oid: '4a245ea3d5e0f947affb4fc65bf4af7a0c708299' },
                    path: 'baz/bonk/quux.ts',
                },
                range,
            }
        )

        assert.deepStrictEqual(location, {
            uri: new URL('git://github.com/foo/bar?4a245ea3d5e0f947affb4fc65bf4af7a0c708299#baz/bonk/quux.ts'),
            range,
        })
    })

    it('falls back to current document', () => {
        const range = scip.Range.fromNumbers(10, 12, 10, 15)

        const location = nodeToLocation(
            {
                uri: 'git://github.com/baz/bonk?c7ad68d72ef7b0d5aac07c22e86fef05d38b06da#source.ts',
            } as sourcegraph.TextDocument,
            {
                resource: {
                    path: 'baz/bonk/quux.ts',
                },
                range,
            }
        )

        assert.deepStrictEqual(location, {
            uri: new URL('git://github.com/baz/bonk?c7ad68d72ef7b0d5aac07c22e86fef05d38b06da#baz/bonk/quux.ts'),
            range,
        })
    })
})
