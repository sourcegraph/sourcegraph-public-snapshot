import { BackendTestContext, filterNodeModules } from './util'
import { createCommit, createLocation } from '../test-utils'

describe('Backend', () => {
    const ctx = new BackendTestContext()
    const repository = 'main'
    const commit = createCommit('0')

    beforeAll(async () => {
        await ctx.init()
        await ctx.convertTestData(repository, commit, '', '/simple/data/main.lsif.gz')
    })

    afterAll(async () => {
        await ctx.teardown()
    })

    it('should find all simple defs of `add` from a.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
            return
        }

        const definitions = await ctx.backend.definitions(repository, commit, 'src/a.ts', { line: 0, character: 17 })
        expect(definitions).toEqual([createLocation('src/a.ts', 0, 16, 0, 19)])
    })

    it('should find all simple defs of `add` from b.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
            return
        }

        const definitions = await ctx.backend.definitions(repository, commit, 'src/b.ts', { line: 2, character: 1 })
        expect(definitions).toEqual([createLocation('src/a.ts', 0, 16, 0, 19)])
    })

    it('should find all simple refs of `add` from a.ts', async () => {
        if (!ctx.backend) {
            fail('failed beforeAll')
            return
        }

        const { locations } = filterNodeModules(
            await ctx.backend.references(repository, commit, 'src/a.ts', { line: 0, character: 17 })
        )

        expect(locations).toContainEqual(createLocation('src/a.ts', 0, 16, 0, 19)) // def
        expect(locations).toContainEqual(createLocation('src/b.ts', 0, 9, 0, 12)) // import
        expect(locations).toContainEqual(createLocation('src/b.ts', 2, 0, 2, 3)) // use
        expect(locations).toContainEqual(createLocation('src/b.ts', 2, 7, 2, 10)) // use
        expect(locations).toContainEqual(createLocation('src/b.ts', 2, 14, 2, 17)) // use
        expect(locations).toHaveLength(5)
    })
})
