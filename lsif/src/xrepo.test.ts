import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { XrepoDatabase, MAX_TRAVERSAL_LIMIT } from './xrepo'
import { createCleanPostgresDatabase, createCommit, truncatePostgresTables } from './test-utils'
import { Connection } from 'typeorm'
import { fail } from 'assert'
import { omit } from 'lodash'

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
        await xrepoDatabase.insertDump('foo', ca, '')
        await xrepoDatabase.insertDump('foo', cc, '')
        await xrepoDatabase.insertDump('foo', cg, '')

        // Test closest commit
        expect((await xrepoDatabase.findClosestDump('foo', ca, 'file'))!.commit).toEqual(ca)
        expect((await xrepoDatabase.findClosestDump('foo', cb, 'file'))!.commit).toEqual(ca)
        expect((await xrepoDatabase.findClosestDump('foo', cc, 'file'))!.commit).toEqual(cc)
        expect((await xrepoDatabase.findClosestDump('foo', cd, 'file'))!.commit).toEqual(cc)
        expect((await xrepoDatabase.findClosestDump('foo', cf, 'file'))!.commit).toEqual(cg)
        expect((await xrepoDatabase.findClosestDump('foo', cg, 'file'))!.commit).toEqual(cg)

        // Multiple nearest are chosen arbitrarily
        expect([ca, cc, cg]).toContain((await xrepoDatabase.findClosestDump('foo', ce, 'file'))!.commit)
        expect([ca, cc]).toContain((await xrepoDatabase.findClosestDump('foo', ch, 'file'))!.commit)
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
        await xrepoDatabase.insertDump('foo', cb, '')

        // Test closest commit
        expect((await xrepoDatabase.findClosestDump('foo', ca, 'file'))!.commit).toEqual(cb)
        expect((await xrepoDatabase.findClosestDump('foo', cb, 'file'))!.commit).toEqual(cb)
        expect((await xrepoDatabase.findClosestDump('foo', cc, 'file'))!.commit).toEqual(cb)
        expect(await xrepoDatabase.findClosestDump('foo', cd, 'file')).toBeUndefined()
        expect(await xrepoDatabase.findClosestDump('foo', ce, 'file')).toBeUndefined()
        expect(await xrepoDatabase.findClosestDump('foo', cf, 'file')).toBeUndefined()
        expect(await xrepoDatabase.findClosestDump('foo', cg, 'file')).toBeUndefined()
        expect(await xrepoDatabase.findClosestDump('foo', ch, 'file')).toBeUndefined()
    })

    it('should return empty string as closest commit with no reachable lsif data', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b]
        //
        // Where LSIF dumps exist at b at roots: root1/ and root2/.

        const ca = createCommit('a')
        const cb = createCommit('b')

        await xrepoDatabase.updateCommits('foo', [[ca, ''], [cb, ca]])

        // Add markers
        await xrepoDatabase.insertDump('foo', cb, 'root1/')
        await xrepoDatabase.insertDump('foo', cb, 'root2/')

        // Test closest commit
        expect(await xrepoDatabase.findClosestDump('foo', ca, 'blah')).toBeUndefined()
        expect(omit(await xrepoDatabase.findClosestDump('foo', cb, 'root1/file.ts'), 'id')).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root1/',
        })
        expect(omit(await xrepoDatabase.findClosestDump('foo', cb, 'root2/file.ts'), 'id')).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root2/',
        })
        expect(omit(await xrepoDatabase.findClosestDump('foo', ca, 'root2/file.ts'), 'id')).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root2/',
        })

        expect(await xrepoDatabase.findClosestDump('foo', ca, 'root3/file.ts')).toBeUndefined()

        await xrepoDatabase.insertDump('foo', cb, '')
        expect(omit(await xrepoDatabase.findClosestDump('foo', ca, 'root2/file.ts'), 'id')).toEqual({
            repository: 'foo',
            commit: cb,
            root: '',
        })
        expect(omit(await xrepoDatabase.findClosestDump('foo', ca, 'root3/file.ts'), 'id')).toEqual({
            repository: 'foo',
            commit: cb,
            root: '',
        })
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
        await xrepoDatabase.insertDump('foo', c0, '')

        // Test closest commit
        expect((await xrepoDatabase.findClosestDump('foo', c0, 'file'))!.commit).toEqual(c0)
        expect((await xrepoDatabase.findClosestDump('foo', c1, 'file'))!.commit).toEqual(c0)
        expect((await xrepoDatabase.findClosestDump('foo', cpen, 'file'))!.commit).toEqual(c0)

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

        expect(await xrepoDatabase.findClosestDump('foo', cmax, 'file')).toBeUndefined()

        // Mark commit 1
        await xrepoDatabase.insertDump('foo', c1, '')

        // Now commit 1 should be found
        expect((await xrepoDatabase.findClosestDump('foo', cmax, 'file'))!.commit).toEqual(c1)
    })
})
