import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { XrepoDatabase, MAX_TRAVERSAL_LIMIT } from './xrepo'
import { createCleanPostgresDatabase, createCommit, truncatePostgresTables } from './test-utils'
import { Connection } from 'typeorm'
import { fail } from 'assert'

describe('XrepoDatabase', () => {
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase

    beforeAll(async () => {
        ;({ connection, cleanup } = await createCleanPostgresDatabase())
        storageRoot = await fs.mkdtemp('xrepo-', { encoding: 'utf8' })
        xrepoDatabase = new XrepoDatabase(connection)
    })

    afterAll(async () => {
        await rmfr(storageRoot)

        if (cleanup) {
            await cleanup()
        }
    })

    beforeEach(async () => {
        if (connection) {
            await truncatePostgresTables(connection)
        }
    })

    it('should find closest commits with LSIF data', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // [a] --+--- b --------+--e -- f --+-- [g]
        //       |              |           |
        //       +-- [c] -- d --+           +--- h

        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')
        const cf = createCommit('f')
        const cg = createCommit('g')
        const ch = createCommit('h')

        await xrepoDatabase.updateCommits('foo', [
            [ca, ''],
            [cb, ca],
            [cc, ca],
            [cd, cc],
            [ce, cb],
            [ce, cd],
            [cf, ce],
            [cg, cf],
            [ch, cf],
        ])

        // Add relations
        await xrepoDatabase.insertDump('foo', ca)
        await xrepoDatabase.insertDump('foo', cc)
        await xrepoDatabase.insertDump('foo', cg)

        // Test closest commit
        expect(await xrepoDatabase.findClosestCommitWithData('foo', ca)).toEqual(ca)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cb)).toEqual(ca)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cc)).toEqual(cc)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cd)).toEqual(cc)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cf)).toEqual(cg)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cg)).toEqual(cg)

        // Multiple nearest are chosen arbitrarily
        expect([ca, cc, cg]).toContain(await xrepoDatabase.findClosestCommitWithData('foo', ce))
        expect([ca, cc]).toContain(await xrepoDatabase.findClosestCommitWithData('foo', ch))
    })

    it('should return empty string as closest commit with no reachable lsif data', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b] ---- c
        //     |
        //     +--- d --+-- e -- f
        //              |
        //              +-- g -- h

        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')
        const cf = createCommit('f')
        const cg = createCommit('g')
        const ch = createCommit('h')

        await xrepoDatabase.updateCommits('foo', [
            [ca, ''],
            [cb, ca],
            [cc, cb],
            [cd, ca],
            [ce, cd],
            [cf, ce],
            [cg, cd],
            [ch, cg],
        ])

        // Add markers
        await xrepoDatabase.insertDump('foo', cb)

        // Test closest commit
        expect(await xrepoDatabase.findClosestCommitWithData('foo', ca)).toEqual(cb)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cb)).toEqual(cb)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cc)).toEqual(cb)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cd)).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', ce)).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cf)).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cg)).toBeUndefined()
        expect(await xrepoDatabase.findClosestCommitWithData('foo', ch)).toBeUndefined()
    })

    it('should not return elements farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the right):
        //
        // 0 -- 1 -- 2 -- ... -- MAX_TRAVERSAL_LIMIT
        //
        // Note: we use '.' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1.1.1...` and `11.11.11...`).

        const c0 = createCommit('0.')
        const c1 = createCommit('1.')
        const cpen = createCommit(`${MAX_TRAVERSAL_LIMIT / 2 - 1}.`)
        const cmax = createCommit(`${MAX_TRAVERSAL_LIMIT / 2}.`)

        const commits: [string, string][] = Array.from({ length: MAX_TRAVERSAL_LIMIT }, (_, i) => [
            createCommit(`${i}.`),
            createCommit(`${i + 1}.`),
        ])

        await xrepoDatabase.updateCommits('foo', commits)

        // Add markers
        await xrepoDatabase.insertDump('foo', c0)

        // Test closest commit
        expect(await xrepoDatabase.findClosestCommitWithData('foo', c0)).toEqual(c0)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', c1)).toEqual(c0)
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cpen)).toEqual(c0)

        // (Assuming MAX_TRAVERSAL_LIMIT = 100)
        // At commit `50`, the traversal limit will be reached before visiting commit `0`
        // because commits are visited in this order:
        //
        // | depth | commit |
        // | ----- | ------ |
        // | 1     | 50     | (with direction 'A')
        // | 2     | 50     | (with direction 'D')
        // | 3     | 51     |
        // | 4     | 49     |
        // | 5     | 52     |
        // | 6     | 48     |
        // | ...   |        |
        // | 99    | 99     |
        // | 100   | 1      | (limit reached)

        expect(await xrepoDatabase.findClosestCommitWithData('foo', cmax)).toBeUndefined()

        // Mark commit 1
        await xrepoDatabase.insertDump('foo', c1)

        // Now commit 1 should be found
        expect(await xrepoDatabase.findClosestCommitWithData('foo', cmax)).toEqual(c1)
    })
})
