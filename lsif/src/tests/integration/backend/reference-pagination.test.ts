import * as util from '../integration-test-util'
import { extractRepos } from './util'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const commit = util.createCommit()

    const ids = {
        a: 100,
        b1: 101,
        b2: 103,
        b3: 104,
        b4: 105,
        b5: 106,
        b6: 107,
        b7: 108,
        b8: 109,
        b9: 110,
        // note: lexiographic order
        b10: 102,
    }

    beforeAll(async () => {
        await ctx.init()
        await Promise.all(
            Object.entries(ids).map(([repositoryName, repositoryId]) =>
                ctx.convertTestData(
                    repositoryId,
                    commit,
                    '',
                    'test',
                    `reference-pagination/data/${repositoryName}.lsif.gz`
                )
            )
        )
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all refs of `add` from repo a', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                await util.queryAllReferences(backend, ids.a, commit, 'src/index.ts', { line: 0, character: 17 }, 50)
            )
        )

        // TODO - test page sizes

        expect(extractRepos(locations)).toEqual([
            ids.a,
            ids.b1,
            ids.b10,
            ids.b2,
            ids.b3,
            ids.b4,
            ids.b5,
            ids.b6,
            ids.b7,
            ids.b8,
            ids.b9,
        ])
    })
})
