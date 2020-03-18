import * as util from '../integration-test-util'
import { lsp } from 'lsif-protocol'
import { MAX_TRAVERSAL_LIMIT } from '../../../shared/constants'
import { extractRepos } from './util'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const repositoryId = 100
    const c0 = util.createCommit(0)
    const c1 = util.createCommit(1)
    const c2 = util.createCommit(2)
    const c3 = util.createCommit(3)
    const c4 = util.createCommit(4)
    const cpen = util.createCommit(MAX_TRAVERSAL_LIMIT * 2 - 1)
    const cmax = util.createCommit(MAX_TRAVERSAL_LIMIT * 2)

    beforeAll(async () => {
        await ctx.init()

        if (!ctx.dumpManager || !ctx.dependencyManager) {
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
                    repositoryId,
                    commit,
                    `${project}/`,
                    'test',
                    `reference-pagination-monorepo/data/${project}${suffix}.lsif.gz`,
                    false
                )
            )
        )

        await ctx.dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>(
                Array.from({ length: MAX_TRAVERSAL_LIMIT * 2 + 1 }, (_, i) => [
                    util.createCommit(i),
                    new Set<string>([util.createCommit(i + 1)]),
                ])
            )
        )
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all refs of `add` within monorepo', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const checkRefs = (locations: lsp.Location[], commit: string, root: string) => {
            expect(locations).toContainEqual(
                util.createLocation(repositoryId, commit, `${root}/src/index.ts`, 0, 9, 0, 12)
            )
            expect(locations).toContainEqual(
                util.createLocation(repositoryId, commit, `${root}/src/index.ts`, 3, 0, 3, 3)
            )
            expect(locations).toContainEqual(
                util.createLocation(repositoryId, commit, `${root}/src/index.ts`, 3, 7, 3, 10)
            )
            expect(locations).toContainEqual(
                util.createLocation(repositoryId, commit, `${root}/src/index.ts`, 3, 14, 3, 17)
            )
            expect(locations).toContainEqual(
                util.createLocation(repositoryId, commit, `${root}/src/index.ts`, 3, 21, 3, 24)
            )
        }

        const testCases = [
            {
                commit: c1,
                defCommit: c1,
                refs: [
                    { root: 'b', commit: c2 },
                    { root: 'e', commit: c0 },
                ],
            },
            {
                commit: c3,
                defCommit: c1,
                refs: [
                    { root: 'b', commit: c2 },
                    { root: 'c', commit: c3 },
                    { root: 'e', commit: c0 },
                ],
            },
            {
                commit: c4,
                defCommit: c1,
                refs: [
                    { root: 'c', commit: c3 },
                    { root: 'e', commit: c0 },
                ],
            },
            {
                commit: cpen,
                defCommit: cpen,
                refs: [{ root: 'f', commit: cmax }],
            },
        ]

        for (const { commit, defCommit, refs } of testCases) {
            const { locations } = util.filterNodeModules(
                util.mapLocations(
                    await util.queryAllReferences(
                        backend,
                        repositoryId,
                        commit,
                        'a/src/index.ts',
                        { line: 0, character: 17 },
                        50
                    )
                )
            )

            expect(locations).toContainEqual(
                util.createLocation(repositoryId, defCommit, 'a/src/index.ts', 0, 16, 0, 19)
            )
            for (const { root, commit: refCommit } of refs) {
                checkRefs(locations, refCommit, root)
            }
            expect(locations).toHaveLength(1 + 5 * refs.length)
        }
    })

    it('should find all refs of `add` from monorepo', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const ids = {
            ext1: 101,
            ext2: 103,
            ext3: 104,
            ext4: 105,
            ext5: 106,
        }

        // Add external references
        await Promise.all(
            Object.values(ids).map(externalRepositoryId =>
                ctx.convertTestData(
                    externalRepositoryId,
                    util.createCommit(0),
                    'f/',
                    'test',
                    'reference-pagination-monorepo/data/f-ref.lsif.gz'
                )
            )
        )

        const fetch = async (
            limit: number
        ): Promise<{ locations: lsp.Location[]; pageSizes: number[]; numPages: number }> =>
            util.filterNodeModules(
                util.mapLocations(
                    await util.queryAllReferences(
                        backend,
                        repositoryId,
                        c3,
                        'a/src/index.ts',
                        { line: 0, character: 17 },
                        limit
                    )
                )
            )

        const ensureSizes = (sizes: number[], expectedSize: number): void => {
            const copy = Array.from(sizes)
            expect(copy.pop()).toBeLessThanOrEqual(expectedSize)
            expect(copy.every(v => v === expectedSize)).toBeTruthy()
        }

        const { locations: locations1, pageSizes: pageSizes1, numPages: numPages1 } = await fetch(1)
        const { locations: locations2, pageSizes: pageSizes2, numPages: numPages2 } = await fetch(5)
        const { locations: locations3, pageSizes: pageSizes3, numPages: numPages3 } = await fetch(100)

        // Ensure we have the correct data (order is asserted here)
        expect(extractRepos(locations1)).toEqual([repositoryId, ids.ext1, ids.ext2, ids.ext3, ids.ext4, ids.ext5])

        // Ensure we have the same data
        expect(locations1).toEqual(locations2)
        expect(locations1).toEqual(locations3)

        // Number of results are the same (no additional duplicates)
        expect(pageSizes1.reduce((a, b) => a + b, 0)).toEqual(pageSizes2.reduce((a, b) => a + b, 0))
        expect(pageSizes1.reduce((a, b) => a + b, 0)).toEqual(pageSizes3.reduce((a, b) => a + b, 0))

        // Ensure pages are full
        ensureSizes(pageSizes1, 1)
        ensureSizes(pageSizes2, 5)
        ensureSizes(pageSizes3, 100)

        // Ensure num pages decrease with page size
        expect(numPages1).toBeGreaterThan(numPages2)
        expect(numPages2).toBeGreaterThan(numPages3)
    })
})
