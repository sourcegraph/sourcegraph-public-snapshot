import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { XrepoDatabase, MAX_TRAVERSAL_LIMIT } from './xrepo'
import { getCleanSqliteDatabase } from './test-utils'
import { entities } from './xrepo.models'

describe('XrepoDatabase', () => {
    let storageRoot!: string

    beforeAll(async () => {
        storageRoot = await fs.mkdtemp('xrepo-', { encoding: 'utf8' })
    })

    afterAll(async () => await rmfr(storageRoot))

    // factory for randomly named xrepo database instance
    const createXrepoDatabase = async () => new XrepoDatabase(await getCleanSqliteDatabase(storageRoot, entities))

    it('should find closest commits with LSIF data', async () => {
        const xrepoDatabase = await createXrepoDatabase()

        // This database has the following commit graph:
        //
        // [a] --+--- b --------+--e -- f --+-- [g]
        //       |              |           |
        //       +-- [c] -- d --+           +--- h

        await xrepoDatabase.updateCommits('foo', [
            ['a', ''],
            ['b', 'a'],
            ['c', 'a'],
            ['d', 'c'],
            ['e', 'b'],
            ['e', 'd'],
            ['f', 'e'],
            ['g', 'f'],
            ['h', 'f'],
        ])

        // Add relations
        await xrepoDatabase.addPackagesAndReferences('foo', 'a', [], [])
        await xrepoDatabase.addPackagesAndReferences('foo', 'c', [], [])
        await xrepoDatabase.addPackagesAndReferences('foo', 'g', [], [])

        // Test closest commit
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'a')).toEqual('a')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'b')).toEqual('a')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'c')).toEqual('c')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'd')).toEqual('c')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'f')).toEqual('g')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'g')).toEqual('g')

        // Multiple nearest are chosen arbitrarily
        expect(['a', 'c', 'g']).toContain(await xrepoDatabase.findClosestCommitWithData('foo', 'e'))
        expect(['a', 'c']).toContain(await xrepoDatabase.findClosestCommitWithData('foo', 'h'))
    })

    it('should return empty string as closest commit with no reachable lsif data', async () => {
        const xrepoDatabase = await createXrepoDatabase()

        // This database has the following commit graph:
        //
        // a --+-- [b] ---- c
        //     |
        //     +--- d --+-- e -- f
        //              |
        //              +-- g -- h

        await xrepoDatabase.updateCommits('foo', [
            ['a', ''],
            ['b', 'a'],
            ['c', 'b'],
            ['d', 'a'],
            ['e', 'd'],
            ['f', 'e'],
            ['g', 'd'],
            ['h', 'g'],
        ])

        // Add markers
        await xrepoDatabase.addPackagesAndReferences('foo', 'b', [], [])

        // Test closest commit
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'a')).toEqual('b')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'b')).toEqual('b')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'c')).toEqual('b')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'd')).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'e')).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'f')).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'g')).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', 'h')).toBeUndefined()
    })

    it('should not return elements farther than MAX_TRAVERSAL_LIMIT', async () => {
        const xrepoDatabase = await createXrepoDatabase()

        // This database has the following commit graph:
        //
        // ... -- (-2) -- (-1) -- [0] -- 1 -- 2 -- ...

        const commits: [string, string][] = []
        for (let i = -(MAX_TRAVERSAL_LIMIT + 3); i < MAX_TRAVERSAL_LIMIT + 3; i++) {
            commits.push([`${i}`, `${i + 1}`])
            commits.push([`${i - 1}`, `${i}`])
        }

        await xrepoDatabase.updateCommits('foo', commits)

        // Add markers
        await xrepoDatabase.addPackagesAndReferences('foo', '0', [], [])

        // Test closest commit
        const limit = MAX_TRAVERSAL_LIMIT
        expect(await xrepoDatabase.findClosestCommitWithData('foo', '0')).toEqual('0')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', '1')).toEqual('0')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', '-1')).toEqual('0')

        for (let i = 0; i <= 2; i++) {
            expect(await xrepoDatabase.findClosestCommitWithData('foo', `${+(limit - i)}`)).toEqual('0')
            expect(await xrepoDatabase.findClosestCommitWithData('foo', `${-(limit - i)}`)).toEqual('0')
        }

        for (let i = 1; i <= 3; i++) {
            expect(await xrepoDatabase.findClosestCommitWithData('foo', `${+(limit + i)}`)).toBeUndefined()
            expect(await xrepoDatabase.findClosestCommitWithData('foo', `${-(limit + i)}`)).toBeUndefined()
        }

        // Modify markers, retest extreme bounds
        await xrepoDatabase.addPackagesAndReferences('foo', '1', [], [])
        expect(await xrepoDatabase.findClosestCommitWithData('foo', `${+(limit + 1)}`)).toEqual('1')
        expect(await xrepoDatabase.findClosestCommitWithData('foo', `${-(limit + 1)}`)).toBeUndefined()
    })
})
