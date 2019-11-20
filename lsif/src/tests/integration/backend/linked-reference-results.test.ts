import * as util from '../integration-test-util'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const repository = 'main'
    const commit = util.createCommit()

    beforeAll(async () => {
        await ctx.init()
        await ctx.convertTestData(repository, commit, '', 'linked-reference-results/data/main.lsif.gz')
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all linked refs of `foo` in index.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const positions = [
            { line: 1, character: 5 },
            { line: 5, character: 5 },
            { line: 9, character: 5 },
            { line: 13, character: 3 },
            { line: 16, character: 3 },
        ]

        for (const position of positions) {
            const { locations } = util.filterNodeModules(
                (await ctx.backend.references(repository, commit, 'src/index.ts', position)) || { locations: [] }
            )

            expect(locations).toContainEqual(util.createLocation('src/index.ts', 1, 4, 1, 7)) // abstract def in I
            expect(locations).toContainEqual(util.createLocation('src/index.ts', 5, 4, 5, 7)) // concrete def in A
            expect(locations).toContainEqual(util.createLocation('src/index.ts', 9, 4, 9, 7)) // concrete def in B
            expect(locations).toContainEqual(util.createLocation('src/index.ts', 13, 2, 13, 5)) // use via I
            expect(locations).toContainEqual(util.createLocation('src/index.ts', 16, 2, 16, 5)) // use via B

            // Ensure no additional references
            expect(locations && locations.length).toEqual(5)
        }
    })
})
