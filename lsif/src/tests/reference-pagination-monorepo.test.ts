import { BackendTestContext, filterNodeModules } from './util'
import { createCommit, createLocation } from '../test-utils'
import { ReferencePaginationContext } from '../backend'
import { MAX_TRAVERSAL_LIMIT } from '../constants'
import { lsp } from 'lsif-protocol'

describe('Backend', () => {
    const ctx = new BackendTestContext()
    const repository = 'monorepo'

    // Note: we use '.' as a suffix for commit numbers on construction so that
    // we can distinguish `1` and `11` (`1.1.1...` and `11.11.11...`).
    const c0 = createCommit('0.')
    const c1 = createCommit('1.')
    const c2 = createCommit('2.')
    const c3 = createCommit('3.')
    const c4 = createCommit('4.')
    const cpen = createCommit(`${MAX_TRAVERSAL_LIMIT * 2 - 1}.`)
    const cmax = createCommit(`${MAX_TRAVERSAL_LIMIT * 2}.`)

    beforeAll(async () => {
        await ctx.init()

        if (!ctx.xrepoDatabase) {
            return
        }

        // The following illustrates the type of dump present at each commit
        // and root intersection. Here, commit `cx` is a parent of `c(x+1)`.
        //
        // -----+---+---+---+---+---+---+
        // --   | a | b | c | d | e | f |
        // -----+---+---+---+---+---+---+
        // c0   |   |   |   | R | R |   |
        // c1   | D |   |   |NoR|   |   |
        // c2   |   | R |NoR|   |   |   |
        // c3   | D |   | R |   |   |   |
        // c4   | D |NoR|   |   |   |   |
        // (...)
        // cpen | D |   |   |   |   |   |
        // cmax |   |   |   |   |   | R |
        //
        // Legend:
        //   - D indicates a dump containing a definition of `add`
        //   - R indicates a dump containing a reference of `add`
        //   - NoR indicates a dump containing no references of `add`

        const dumps = [
            { commit: c0, project: 'd', suffix: '-ref' },
            { commit: c0, project: 'e', suffix: '-ref' },
            { commit: c1, project: 'a', suffix: '' },
            { commit: c1, project: 'd', suffix: '-noref' },
            { commit: c2, project: 'b', suffix: '-ref' },
            { commit: c2, project: 'c', suffix: '-noref' },
            { commit: c3, project: 'c', suffix: '-ref' },
            { commit: c4, project: 'b', suffix: '-noref' },
            { commit: cpen, project: 'a', suffix: '' },
            { commit: cmax, project: 'f', suffix: '-ref' },
        ]

        await Promise.all(
            dumps.map(({ commit, project, suffix }) =>
                ctx.convertTestData(
                    repository,
                    commit,
                    `${project}/`,
                    `reference-pagination-monorepo/data/${project}${suffix}.lsif.gz`,
                    false
                )
            )
        )

        await ctx.xrepoDatabase.updateCommits(
            repository,
            Array.from({ length: MAX_TRAVERSAL_LIMIT * 2 + 1 }, (_, i) => [
                createCommit(`${i}.`),
                createCommit(`${i + 1}.`),
            ])
        )
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all refs of `add` within monorepo', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
            return
        }

        const checkRefs = (locations: lsp.Location[], root: string) => {
            expect(locations).toContainEqual(createLocation(`${root}/src/index.ts`, 0, 9, 0, 12))
            expect(locations).toContainEqual(createLocation(`${root}/src/index.ts`, 3, 0, 3, 3))
            expect(locations).toContainEqual(createLocation(`${root}/src/index.ts`, 3, 7, 3, 10))
            expect(locations).toContainEqual(createLocation(`${root}/src/index.ts`, 3, 14, 3, 17))
            expect(locations).toContainEqual(createLocation(`${root}/src/index.ts`, 3, 21, 3, 24))
        }

        const testCases = [
            { commit: c1, refs: ['b', 'e'] },
            { commit: c3, refs: ['b', 'c', 'e'] },
            { commit: c4, refs: ['c', 'e'] },
            { commit: cpen, refs: ['f'] },
        ]

        for (const { commit, refs } of testCases) {
            const fetch = async (paginationContext?: ReferencePaginationContext) =>
                filterNodeModules(
                    await backend.references(
                        repository,
                        commit,
                        'a/src/index.ts',
                        {
                            line: 0,
                            character: 17,
                        },
                        paginationContext
                    )
                )

            // TODO - test pagination as well
            const { locations, cursor } = await fetch()
            expect(cursor).toBeUndefined()

            expect(locations).toContainEqual(createLocation('a/src/index.ts', 0, 16, 0, 19))
            for (const root of refs) {
                checkRefs(locations, root)
            }
            expect(locations).toHaveLength(1 + 5 * refs.length)
        }
    })
})
