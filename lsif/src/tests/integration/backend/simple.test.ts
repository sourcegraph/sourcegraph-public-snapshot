import * as util from '../integration-test-util'

describe('Backend', () => {
    const ctx = new util.BackendTestContext()
    const repositoryId = 100
    const repositoryName = 'main'
    const commit = util.createCommit()

    beforeAll(async () => {
        await ctx.init()
        await ctx.convertTestData(repositoryId, repositoryName, commit, '', '/simple/data/main.lsif.gz')
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all simple defs of `add` from a.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const definitions = await ctx.backend.definitions(repositoryId, repositoryName, commit, 'src/a.ts', {
            line: 0,
            character: 17,
        })
        expect(definitions?.map(util.mapLocation)).toEqual([
            util.createLocation(repositoryId, commit, 'src/a.ts', 0, 16, 0, 19),
        ])
    })

    it('should find all simple defs of `add` from b.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const definitions = await ctx.backend.definitions(repositoryId, repositoryName, commit, 'src/b.ts', {
            line: 2,
            character: 1,
        })
        expect(definitions?.map(util.mapLocation)).toEqual([
            util.createLocation(repositoryId, commit, 'src/a.ts', 0, 16, 0, 19),
        ])
    })

    it('should find all simple refs of `add` from a.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
        }

        const { locations } = util.filterNodeModules(
            util.mapLocations(
                (await ctx.backend.references(repositoryId, repositoryName, commit, 'src/a.ts', {
                    line: 0,
                    character: 17,
                })) || {
                    locations: [],
                }
            )
        )

        expect(locations).toContainEqual(util.createLocation(repositoryId, commit, 'src/a.ts', 0, 16, 0, 19)) // def
        expect(locations).toContainEqual(util.createLocation(repositoryId, commit, 'src/b.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(util.createLocation(repositoryId, commit, 'src/b.ts', 2, 0, 2, 3)) // use
        expect(locations).toContainEqual(util.createLocation(repositoryId, commit, 'src/b.ts', 2, 7, 2, 10)) // use
        expect(locations).toContainEqual(util.createLocation(repositoryId, commit, 'src/b.ts', 2, 14, 2, 17)) // use
        expect(locations).toHaveLength(5)
    })
})
