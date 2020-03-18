import * as util from '../integration-test-util'
import { extractRepos } from './util'
import { lsp } from 'lsif-protocol'
import { Backend } from '../../../server/backend/backend'

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

    const fetch = async (
        backend: Backend,
        limit: number,
        remoteDumpLimit?: number
    ): Promise<{ locations: lsp.Location[]; pageSizes: number[]; numPages: number }> =>
        util.filterNodeModules(
            util.mapLocations(
                await util.queryAllReferences(
                    backend,
                    ids.a,
                    commit,
                    'src/index.ts',
                    { line: 0, character: 17 },
                    limit,
                    remoteDumpLimit
                )
            )
        )

    const ensureSizes = (sizes: number[], expectedSize: number): void => {
        const copy = Array.from(sizes)
        expect(copy.pop()).toBeLessThanOrEqual(expectedSize)
        expect(copy.every(v => v === expectedSize)).toBeTruthy()
    }

    it('should find all refs of `add` from repo a', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const { locations: locations1, pageSizes: pageSizes1, numPages: numPages1 } = await fetch(backend, 1)
        const { locations: locations2, pageSizes: pageSizes2, numPages: numPages2 } = await fetch(backend, 5)
        const { locations: locations3, pageSizes: pageSizes3, numPages: numPages3 } = await fetch(backend, 100)

        // Ensure we have the correct data (order is asserted here)
        expect(extractRepos(locations1)).toEqual([
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

    it('should not be externally affected by remote dump limit', async () => {
        const backend = ctx.backend
        if (!backend) {
            fail('failed beforeAll')
        }

        const { locations: locations1, pageSizes: pageSizes1 } = await fetch(backend, 5, 1)
        const { locations: locations2, pageSizes: pageSizes2 } = await fetch(backend, 5, 5)
        const { locations: locations3, pageSizes: pageSizes3 } = await fetch(backend, 5, 10)

        // Ensure we have the same data
        expect(locations1).toEqual(locations2)
        expect(locations1).toEqual(locations3)

        // Ensure we have the same page distribution
        expect(pageSizes1).toEqual(pageSizes2)
        expect(pageSizes1).toEqual(pageSizes3)
    })
})
