import * as util from '../integration-test-util'
import { lsp } from 'lsif-protocol'
import { ReferencePaginationContext } from '../../../server/backend/backend'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const commit = util.createCommit()

    beforeAll(async () => {
        await ctx.init()
        await Promise.all(
            ['a', 'b1', 'b2', 'b3', 'b4', 'b5', 'b6', 'b7', 'b8', 'b9', 'b10'].map(r =>
                ctx.convertTestData(r, commit, '', `reference-pagination/data/${r}.lsif.gz`)
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
                (await backend.references(
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

        const extractRepos = (references: lsp.Location[]): string[] =>
            // extract the repo name from git://{repo}?{commit}#{path}, or return '' (indicating a local repo)
            Array.from(new Set(references.map(r => (r.uri.match(/git:\/\/([^?]+)\?.+/) || ['', ''])[1]))).sort()

        // Ensure paging gets us expected results per page
        expect(extractRepos(locations1)).toEqual(['', 'b1', 'b10', 'b2'])
        expect(extractRepos(locations2)).toEqual(['b3', 'b4', 'b5'])
        expect(extractRepos(locations3)).toEqual(['b6', 'b7', 'b8'])
        expect(extractRepos(locations4)).toEqual(['b9'])
    })
})
