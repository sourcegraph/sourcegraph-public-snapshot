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

            const { locations, cursor } = await fetch()
            expect(cursor).toBeUndefined()

            expect(locations).toContainEqual(createLocation('a/src/index.ts', 0, 16, 0, 19))
            for (const root of refs) {
                checkRefs(locations, root)
            }
            expect(locations).toHaveLength(1 + 5 * refs.length)
        }
    })

    it('should find all refs of `add` from monorepo', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
            return
        }

        // Add external references
        const repos = ['ext1', 'ext2', 'ext3', 'ext4', 'ext5']
        const filename = 'reference-pagination-monorepo/data/f-ref.lsif.gz'
        await Promise.all(repos.map(r => ctx.convertTestData(r, createCommit(r), 'f/', filename)))

        const fetch = async (paginationContext?: ReferencePaginationContext) =>
            filterNodeModules(
                await backend.references(
                    repository,
                    c3,
                    'a/src/index.ts',
                    {
                        line: 0,
                        character: 17,
                    },
                    paginationContext
                )
            )

        const { locations: locations0, cursor: cursor0 } = await fetch({ limit: 50 }) // all local
        const { locations: locations1, cursor: cursor1 } = await fetch({ limit: 50, cursor: cursor0 }) // all remote

        const { locations: locations2, cursor: cursor2 } = await fetch({ limit: 2 }) // b, c
        const { locations: locations3, cursor: cursor3 } = await fetch({ limit: 2, cursor: cursor2 }) // e
        const { locations: locations4, cursor: cursor4 } = await fetch({ limit: 2, cursor: cursor3 }) // ext1, ext2
        const { locations: locations5, cursor: cursor5 } = await fetch({ limit: 2, cursor: cursor4 }) // ext3, ext4
        const { locations: locations6, cursor: cursor6 } = await fetch({ limit: 2, cursor: cursor5 }) // ext5

        // Ensure paging through sets of results gets us everything
        expect(locations0).toEqual(locations2.concat(...locations3))
        expect(locations1).toEqual(locations4.concat(...locations5, ...locations6))

        // Ensure cursor is not provided at the end of a set of results
        expect(cursor1).toBeUndefined()
        expect(cursor6).toBeUndefined()

        const extractRepos = (references: lsp.Location[]): string[] =>
            // extract the repo name from git://{repo}?{commit}#{path}, or return '' (indicating a local repo)
            Array.from(new Set(references.map(r => (r.uri.match(/git:\/\/([^?]+)\?.+/) || ['', ''])[1]))).sort()

        // Ensure paging gets us expected results per page
        expect(extractRepos(locations0)).toEqual([''])
        expect(extractRepos(locations1)).toEqual(['ext1', 'ext2', 'ext3', 'ext4', 'ext5'])
        expect(extractRepos(locations2)).toEqual([''])
        expect(extractRepos(locations3)).toEqual([''])
        expect(extractRepos(locations4)).toEqual(['ext1', 'ext2'])
        expect(extractRepos(locations5)).toEqual(['ext3', 'ext4'])
        expect(extractRepos(locations6)).toEqual(['ext5'])
    })
})
