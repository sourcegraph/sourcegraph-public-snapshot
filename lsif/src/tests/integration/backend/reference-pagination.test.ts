import * as util from '../integration-test-util'
import { ReferencePaginationContext } from '../../../server/backend/backend'
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
                    repositoryName,
                    commit,
                    '',
                    `reference-pagination/data/${repositoryName}.lsif.gz`
                )
            )
        )
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all defs of `add` from repo a', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const fetch = async (paginationContext?: ReferencePaginationContext) =>
            util.filterNodeModules(
                util.mapLocations(
                    (await backend.references(
                        ids.a,
                        'a',
                        commit,
                        'src/index.ts',
                        {
                            line: 0,
                            character: 17,
                        },
                        paginationContext
                    )) || { locations: [] }
                )
            )

        const { locations: locations0, cursor: cursor0 } = await fetch() // everything
        const { locations: locations1, cursor: cursor1 } = await fetch({ limit: 3 }) // a, b1, b10, b2
        const { locations: locations2, cursor: cursor2 } = await fetch({ limit: 3, cursor: cursor1 }) // b3, b4, b5
        const { locations: locations3, cursor: cursor3 } = await fetch({ limit: 3, cursor: cursor2 }) // b6, b7, b8
        const { locations: locations4, cursor: cursor4 } = await fetch({ limit: 3, cursor: cursor3 }) // b9

        // Ensure paging through sets of results gets us everything
        expect(locations0).toEqual(locations1.concat(...locations2, ...locations3, ...locations4))

        // Ensure cursor is not provided at the end of a set of results
        expect(cursor0).toBeUndefined()
        expect(cursor4).toBeUndefined()

        // Ensure paging gets us expected results per page
        expect(extractRepos(locations1)).toEqual([ids.a, ids.b1, ids.b10, ids.b2])
        expect(extractRepos(locations2)).toEqual([ids.b3, ids.b4, ids.b5])
        expect(extractRepos(locations3)).toEqual([ids.b6, ids.b7, ids.b8])
        expect(extractRepos(locations4)).toEqual([ids.b9])
    })
})
