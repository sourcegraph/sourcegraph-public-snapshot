import * as util from '../integration-test-util'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const commit = util.createCommit()

    const ids = {
        a: 100,
        b1: 101,
        b2: 103,
        b3: 104,
        c1: 105,
        c2: 106,
        c3: 107,
    }

    beforeAll(async () => {
        await ctx.init()
        await Promise.all(
            Object.entries(ids).map(([repositoryName, repositoryId]) =>
                ctx.convertTestData(repositoryId, commit, '', `xrepo/data/${repositoryName}.lsif.gz`)
            )
        )
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all cross-repo defs of `add` from repo a', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const definitions = await ctx.backend.definitions(ids.a, commit, 'src/index.ts', {
            line: 11,
            character: 18,
        })
        expect(definitions?.map(util.mapLocation)).toEqual([
            util.createLocation(ids.a, commit, 'src/index.ts', 0, 16, 0, 19),
        ])
    })

    it('should find all cross-repo defs of `add` from repo b1', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const definitions = await ctx.backend.definitions(ids.b1, commit, 'src/index.ts', {
            line: 3,
            character: 12,
        })
        expect(definitions?.map(util.mapLocation)).toEqual([
            util.createLocation(ids.a, commit, 'src/index.ts', 0, 16, 0, 19),
        ])
    })

    it('should find all cross-repo defs of `mul` from repo b1', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const definitions = await ctx.backend.definitions(ids.b1, commit, 'src/index.ts', {
            line: 3,
            character: 16,
        })
        expect(definitions?.map(util.mapLocation)).toEqual([
            util.createLocation(ids.a, commit, 'src/index.ts', 4, 16, 4, 19),
        ])
    })

    it('should find all cross-repo refs of `mul` from repo a', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                (await ctx.backend.references(ids.a, util.createCommit(0), 'src/index.ts', {
                    line: 4,
                    character: 19,
                })) || { locations: [] }
            )
        )

        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 4, 16, 4, 19)) // def
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(locations?.length).toEqual(10)
    })

    it('should find all cross-repo refs of `mul` from repo b1', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                (await ctx.backend.references(ids.b1, util.createCommit(0), 'src/index.ts', {
                    line: 3,
                    character: 16,
                })) || { locations: [] }
            )
        )

        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 4, 16, 4, 19)) // def
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 0, 14, 0, 17)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(locations?.length).toEqual(10)
    })

    it('should find all cross-repo refs of `add` from repo a', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                (await ctx.backend.references(ids.a, util.createCommit(0), 'src/index.ts', {
                    line: 0,
                    character: 17,
                })) || { locations: [] }
            )
        )

        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 0, 16, 0, 19)) // def
        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(locations?.length).toEqual(20)
    })

    it('should find all cross-repo refs of `add` from repo c1', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                (await ctx.backend.references(ids.c1, util.createCommit(0), 'src/index.ts', {
                    line: 3,
                    character: 16,
                })) || { locations: [] }
            )
        )

        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 0, 16, 0, 19)) // def
        expect(locations).toContainEqual(util.createLocation(ids.a, commit, 'src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b1, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b2, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.b3, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c1, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c2, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(locations).toContainEqual(util.createLocation(ids.c3, commit, 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(locations?.length).toEqual(20)
    })
})
