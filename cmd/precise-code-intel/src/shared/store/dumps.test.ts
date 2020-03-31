import * as util from '../test-util'
import * as pgModels from '../models/pg'
import nock from 'nock'
import { Connection } from 'typeorm'
import { DumpManager } from './dumps'
import { fail } from 'assert'
import { MAX_TRAVERSAL_LIMIT } from '../constants'

describe('DumpManager', () => {
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let dumpManager!: DumpManager

    let counter = 100
    const nextId = () => {
        counter++
        return counter
    }

    beforeAll(async () => {
        ;({ connection, cleanup } = await util.createCleanPostgresDatabase())
        dumpManager = new DumpManager(connection)
    })

    afterAll(async () => {
        if (cleanup) {
            await cleanup()
        }
    })

    beforeEach(async () => {
        if (connection) {
            await util.truncatePostgresTables(connection)
        }
    })

    it('should find closest commits with LSIF data (first commit graph)', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // [a] --+--- b --------+--e -- f --+-- [g]
        //       |              |           |
        //       +-- [c] -- d --+           +--- h

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()
        const ch = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([ca])],
                [cd, new Set([cc])],
                [ce, new Set([cb])],
                [ce, new Set([cd])],
                [cf, new Set([ce])],
                [cg, new Set([cf])],
                [ch, new Set([cf])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, ca, '', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cc, '', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cg, '', 'test')

        const d1 = await dumpManager.findClosestDumps(repositoryId, ca, 'file.ts')
        const d2 = await dumpManager.findClosestDumps(repositoryId, cb, 'file.ts')
        const d3 = await dumpManager.findClosestDumps(repositoryId, cc, 'file.ts')
        const d4 = await dumpManager.findClosestDumps(repositoryId, cd, 'file.ts')
        const d5 = await dumpManager.findClosestDumps(repositoryId, cf, 'file.ts')
        const d6 = await dumpManager.findClosestDumps(repositoryId, cg, 'file.ts')
        const d7 = await dumpManager.findClosestDumps(repositoryId, ce, 'file.ts')
        const d8 = await dumpManager.findClosestDumps(repositoryId, ch, 'file.ts')

        expect(d1).toHaveLength(1)
        expect(d2).toHaveLength(1)
        expect(d3).toHaveLength(1)
        expect(d4).toHaveLength(1)
        expect(d5).toHaveLength(1)
        expect(d6).toHaveLength(1)
        expect(d7).toHaveLength(1)
        expect(d8).toHaveLength(1)

        // Test closest commit
        expect(d1[0].commit).toEqual(ca)
        expect(d2[0].commit).toEqual(ca)
        expect(d3[0].commit).toEqual(cc)
        expect(d4[0].commit).toEqual(cc)
        expect(d5[0].commit).toEqual(cg)
        expect(d6[0].commit).toEqual(cg)

        // Multiple nearest are chosen arbitrarily
        expect([ca, cc, cg]).toContain(d7[0].commit)
        expect([ca, cc]).toContain(d8[0].commit)
    })

    it('should find closest commits with LSIF data (second commit graph)', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b] ---- c
        //     |
        //     +--- d --+-- e -- f
        //              |
        //              +-- g -- h

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()
        const ch = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([ca])],
                [ce, new Set([cd])],
                [cf, new Set([ce])],
                [cg, new Set([cd])],
                [ch, new Set([cg])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, cb, '', 'test')

        const d1 = await dumpManager.findClosestDumps(repositoryId, ca, 'file.ts')
        const d2 = await dumpManager.findClosestDumps(repositoryId, cb, 'file.ts')
        const d3 = await dumpManager.findClosestDumps(repositoryId, cc, 'file.ts')

        expect(d1).toHaveLength(1)
        expect(d2).toHaveLength(1)
        expect(d3).toHaveLength(1)

        // Test closest commit
        expect(d1[0].commit).toEqual(cb)
        expect(d2[0].commit).toEqual(cb)
        expect(d3[0].commit).toEqual(cb)
        expect(await dumpManager.findClosestDumps(repositoryId, cd, 'file.ts')).toHaveLength(0)
        expect(await dumpManager.findClosestDumps(repositoryId, ce, 'file.ts')).toHaveLength(0)
        expect(await dumpManager.findClosestDumps(repositoryId, cf, 'file.ts')).toHaveLength(0)
        expect(await dumpManager.findClosestDumps(repositoryId, cg, 'file.ts')).toHaveLength(0)
        expect(await dumpManager.findClosestDumps(repositoryId, ch, 'file.ts')).toHaveLength(0)
    })

    it('should find closest commits with LSIF data (distinct roots)', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b]
        //
        // Where LSIF dumps exist at b at roots: root1/ and root2/.

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'root1/', '')
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'root2/', '')

        // Test closest commit
        const d1 = await dumpManager.findClosestDumps(repositoryId, ca, 'blah')
        const d2 = await dumpManager.findClosestDumps(repositoryId, cb, 'root1/file.ts')
        const d3 = await dumpManager.findClosestDumps(repositoryId, cb, 'root2/file.ts')
        const d4 = await dumpManager.findClosestDumps(repositoryId, ca, 'root2/file.ts')

        expect(d1).toHaveLength(0)
        expect(d2).toHaveLength(1)
        expect(d3).toHaveLength(1)
        expect(d4).toHaveLength(1)

        expect(d2[0].commit).toEqual(cb)
        expect(d2[0].root).toEqual('root1/')
        expect(d3[0].commit).toEqual(cb)
        expect(d3[0].root).toEqual('root2/')
        expect(d4[0].commit).toEqual(cb)
        expect(d4[0].root).toEqual('root2/')

        const d5 = await dumpManager.findClosestDumps(repositoryId, ca, 'root3/file.ts')
        expect(d5).toHaveLength(0)

        await util.insertDump(connection, dumpManager, repositoryId, cb, '', '')
        const d6 = await dumpManager.findClosestDumps(repositoryId, ca, 'root2/file.ts')
        const d7 = await dumpManager.findClosestDumps(repositoryId, ca, 'root3/file.ts')

        expect(d6).toHaveLength(1)
        expect(d7).toHaveLength(1)

        expect(d6[0].commit).toEqual(cb)
        expect(d6[0].root).toEqual('')
        expect(d7[0].commit).toEqual(cb)
        expect(d7[0].root).toEqual('')
    })

    it('should find closest commits with LSIF data (overlapping roots)', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a -- b --+-- c --+-- e -- f
        //          |       |
        //          +-- d --+
        //
        // With the following LSIF dumps:
        //
        // | Commit | Root    | Indexer |
        // | ------ + ------- + ------- |
        // | a      | root3/  | A       |
        // | a      | root4/  | B       |
        // | b      | root1/  | A       |
        // | b      | root2/  | A       |
        // | b      |         | B       | (overwrites root4/ at commit a)
        // | c      | root1/  | A       | (overwrites root1/ at commit b)
        // | d      |         | B       | (overwrites (root) at commit b)
        // | e      | root2/  | A       | (overwrites root2/ at commit b)
        // | f      | root1/  | A       | (overwrites root1/ at commit b)

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([cb])],
                [ce, new Set([cc, cd])],
                [cf, new Set([ce])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, ca, 'root3/', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, ca, 'root4/', 'B')
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'root1/', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'root2/', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, cb, '', 'B')
        await util.insertDump(connection, dumpManager, repositoryId, cc, 'root1/', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, cd, '', 'B')
        await util.insertDump(connection, dumpManager, repositoryId, ce, 'root2/', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, cf, 'root1/', 'A')

        // Test closest commit
        const d1 = await dumpManager.findClosestDumps(repositoryId, cd, 'root1/file.ts')
        expect(d1).toHaveLength(2)
        expect(d1[0].commit).toEqual(cd)
        expect(d1[0].root).toEqual('')
        expect(d1[0].indexer).toEqual('B')
        expect(d1[1].commit).toEqual(cb)
        expect(d1[1].root).toEqual('root1/')
        expect(d1[1].indexer).toEqual('A')

        const d2 = await dumpManager.findClosestDumps(repositoryId, ce, 'root2/file.ts')
        expect(d2).toHaveLength(2)
        expect(d2[0].commit).toEqual(ce)
        expect(d2[0].root).toEqual('root2/')
        expect(d2[0].indexer).toEqual('A')
        expect(d2[1].commit).toEqual(cd)
        expect(d2[1].root).toEqual('')
        expect(d2[1].indexer).toEqual('B')

        const d3 = await dumpManager.findClosestDumps(repositoryId, cc, 'root3/file.ts')
        expect(d3).toHaveLength(2)
        expect(d3[0].commit).toEqual(cb)
        expect(d3[0].root).toEqual('')
        expect(d3[0].indexer).toEqual('B')
        expect(d3[1].commit).toEqual(ca)
        expect(d3[1].root).toEqual('root3/')
        expect(d3[1].indexer).toEqual('A')

        const d4 = await dumpManager.findClosestDumps(repositoryId, ca, 'root4/file.ts')
        expect(d4).toHaveLength(1)
        expect(d4[0].commit).toEqual(ca)
        expect(d4[0].root).toEqual('root4/')
        expect(d4[0].indexer).toEqual('B')

        const d5 = await dumpManager.findClosestDumps(repositoryId, cb, 'root4/file.ts')
        expect(d5).toHaveLength(1)
        expect(d5[0].commit).toEqual(cb)
        expect(d5[0].root).toEqual('')
        expect(d5[0].indexer).toEqual('B')
    })

    it('should not return elements farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the left):
        //
        // MAX_TRAVERSAL_LIMIT -- ... -- 2 -- 1 -- 0
        //
        // Note: we use 'a' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1a1a1a...` and `11a11a11a..`).

        const repositoryId = nextId()
        const c0 = util.createCommit(0)
        const c1 = util.createCommit(1)
        const cpen = util.createCommit(MAX_TRAVERSAL_LIMIT / 2 - 1)
        const cmax = util.createCommit(MAX_TRAVERSAL_LIMIT / 2)

        const commits = new Map<string, Set<string>>(
            Array.from({ length: MAX_TRAVERSAL_LIMIT }, (_, i) => [
                util.createCommit(i),
                new Set([util.createCommit(i + 1)]),
            ])
        )

        // Add relations
        await dumpManager.updateCommits(repositoryId, commits)

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, c0, '', 'test')

        const d1 = await dumpManager.findClosestDumps(repositoryId, c0, 'file.ts')
        const d2 = await dumpManager.findClosestDumps(repositoryId, c1, 'file.ts')
        const d3 = await dumpManager.findClosestDumps(repositoryId, cpen, 'file.ts')

        expect(d1).toHaveLength(1)
        expect(d2).toHaveLength(1)
        expect(d3).toHaveLength(1)

        // Test closest commit
        expect(d1[0].commit).toEqual(c0)
        expect(d2[0].commit).toEqual(c0)
        expect(d3[0].commit).toEqual(c0)

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

        expect(await dumpManager.findClosestDumps(repositoryId, cmax, 'file.ts')).toHaveLength(0)

        // Add closer dump
        await util.insertDump(connection, dumpManager, repositoryId, c1, '', 'test')

        // Now commit 1 should be found
        const dumps = await dumpManager.findClosestDumps(repositoryId, cmax, 'file.ts')
        expect(dumps).toHaveLength(1)
        expect(dumps[0].commit).toEqual(c1)
    })

    it('should prune overlapping roots during visibility check', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a -- b -- c -- d -- e -- f -- g

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([cc])],
                [ce, new Set([cd])],
                [cf, new Set([ce])],
                [cg, new Set([cf])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, ca, 'r1', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'r2', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cc, '', 'test') // overwrites r1, r2
        const d1 = await util.insertDump(connection, dumpManager, repositoryId, cd, 'r3', 'test') // overwrites ''
        const d2 = await util.insertDump(connection, dumpManager, repositoryId, cf, 'r4', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cg, 'r5', 'test') // not traversed

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, cf)
        const visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([d1.id, d2.id])
    })

    it('should prune overlapping roots of the same indexer during visibility check', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }
        // This database has the following commit graph:
        //
        // a -- b -- c -- d -- e -- f -- g

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([cc])],
                [ce, new Set([cd])],
                [cf, new Set([ce])],
                [cg, new Set([cf])],
            ])
        )

        // Add dumps from indexer A
        await util.insertDump(connection, dumpManager, repositoryId, ca, 'r1', 'A')
        const d1 = await util.insertDump(connection, dumpManager, repositoryId, cc, 'r2', 'A')
        const d2 = await util.insertDump(connection, dumpManager, repositoryId, cd, 'r1', 'A') // overwrites r1
        const d3 = await util.insertDump(connection, dumpManager, repositoryId, cf, 'r3', 'A')
        await util.insertDump(connection, dumpManager, repositoryId, cg, 'r4', 'A') // not traversed

        // Add dumps from indexer B
        await util.insertDump(connection, dumpManager, repositoryId, ca, 'r1', 'B')
        await util.insertDump(connection, dumpManager, repositoryId, cc, 'r2', 'B')
        await util.insertDump(connection, dumpManager, repositoryId, cd, '', 'B') // overwrites r1, r2
        const d5 = await util.insertDump(connection, dumpManager, repositoryId, ce, 'r3', 'B') // overwrites ''

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, cf)
        const visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([d1.id, d2.id, d3.id, d5.id])
    })

    it('should traverse branching paths during visibility check', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b] --- c ---+
        //     |                |
        //     +--- d --- [e] --+ -- [h] --+-- [i]
        //     |                           |
        //     +-- [f] --- g --------------+

        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const ch = util.createCommit()
        const ci = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()

        // Add relations
        await dumpManager.updateCommits(
            repositoryId,
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([ca])],
                [ce, new Set([cd])],
                [ch, new Set([cc, ce])],
                [ci, new Set([ch, cg])],
                [cf, new Set([ca])],
                [cg, new Set([cf])],
            ])
        )

        // Add dumps
        await util.insertDump(connection, dumpManager, repositoryId, cb, 'r2', 'test')
        const dump1 = await util.insertDump(connection, dumpManager, repositoryId, ce, 'r2/a', 'test') // overwrites r2 in commit b
        const dump2 = await util.insertDump(connection, dumpManager, repositoryId, ce, 'r2/b', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cf, 'r1/a', 'test')
        await util.insertDump(connection, dumpManager, repositoryId, cf, 'r1/b', 'test')
        const dump3 = await util.insertDump(connection, dumpManager, repositoryId, ch, 'r1', 'test') // overwrites r1/{a,b} in commit f
        const dump4 = await util.insertDump(connection, dumpManager, repositoryId, ci, 'r3', 'test')

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, ci)
        const visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([
            dump1.id,
            dump2.id,
            dump3.id,
            dump4.id,
        ])
    })

    it('should not set dumps visible farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!dumpManager) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the left):
        //
        // (MAX_TRAVERSAL_LIMIT + 1) -- ... -- 2 -- 1 -- 0
        //
        // Note: we use 'a' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1a1a1a...` and `11a11a11a...`).

        const repositoryId = nextId()
        const c0 = util.createCommit(0)
        const c1 = util.createCommit(1)
        const cpen = util.createCommit(MAX_TRAVERSAL_LIMIT - 1)
        const cmax = util.createCommit(MAX_TRAVERSAL_LIMIT)

        const commits = new Map<string, Set<string>>(
            Array.from({ length: MAX_TRAVERSAL_LIMIT + 1 }, (_, i) => [
                util.createCommit(i),
                new Set([util.createCommit(i + 1)]),
            ])
        )

        // Add relations
        await dumpManager.updateCommits(repositoryId, commits)

        // Add dumps
        const dump1 = await util.insertDump(connection, dumpManager, repositoryId, cmax, '', 'test')

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, cmax)
        let visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, c1)
        visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await dumpManager.updateDumpsVisibleFromTip(repositoryId, c0)
        visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([])

        // Add closer dump
        const dump2 = await util.insertDump(connection, dumpManager, repositoryId, cpen, '', 'test')

        // Now commit cpen should be found
        await dumpManager.updateDumpsVisibleFromTip(repositoryId, c0)
        visibleDumps = await dumpManager.getVisibleDumps(repositoryId)
        expect(visibleDumps.map((dump: pgModels.LsifDump) => dump.id).sort()).toEqual([dump2.id])
    })
})

describe('discoverAndUpdateCommit', () => {
    let counter = 200
    const nextId = () => {
        counter++
        return counter
    }

    it('should update tracked commits', async () => {
        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()

        nock('http://frontend')
            .post(`/.internal/git/${repositoryId}/exec`)
            .reply(200, `${ca}\n${cb} ${ca}\n${cc} ${cb}`)

        const { connection, cleanup } = await util.createCleanPostgresDatabase()

        try {
            const dumpManager = new DumpManager(connection)
            await util.insertDump(connection, dumpManager, repositoryId, ca, '', 'test')

            await dumpManager.updateCommits(
                repositoryId,
                await dumpManager.discoverCommits({
                    repositoryId,
                    commit: cc,
                    frontendUrl: 'frontend',
                })
            )

            // Ensure all commits are now tracked
            expect((await connection.getRepository(pgModels.Commit).find()).map(c => c.commit).sort()).toEqual([
                ca,
                cb,
                cc,
            ])
        } finally {
            await cleanup()
        }
    })

    it('should early-out if commit is tracked', async () => {
        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()

        const { connection, cleanup } = await util.createCleanPostgresDatabase()

        try {
            const dumpManager = new DumpManager(connection)
            await util.insertDump(connection, dumpManager, repositoryId, ca, '', 'test')
            await dumpManager.updateCommits(
                repositoryId,
                new Map<string, Set<string>>([[cb, new Set()]])
            )

            // This test ensures the following does not make a gitserver request.
            // As we did not register a nock interceptor, any request will result
            // in an exception being thrown.

            await dumpManager.updateCommits(
                repositoryId,
                await dumpManager.discoverCommits({
                    repositoryId,
                    commit: cb,
                    frontendUrl: 'frontend',
                })
            )
        } finally {
            await cleanup()
        }
    })

    it('should early-out if repository is unknown', async () => {
        const repositoryId = nextId()
        const ca = util.createCommit()

        const { connection, cleanup } = await util.createCleanPostgresDatabase()

        try {
            const dumpManager = new DumpManager(connection)

            // This test ensures the following does not make a gitserver request.
            // As we did not register a nock interceptor, any request will result
            // in an exception being thrown.

            await dumpManager.updateCommits(
                repositoryId,
                await dumpManager.discoverCommits({
                    repositoryId,
                    commit: ca,
                    frontendUrl: 'frontend',
                })
            )
        } finally {
            await cleanup()
        }
    })
})

describe('discoverAndUpdateTips', () => {
    let counter = 300
    const nextId = () => {
        counter++
        return counter
    }

    it('should update tips', async () => {
        const repositoryId = nextId()
        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()

        nock('http://frontend')
            .post(`/.internal/git/${repositoryId}/exec`, { args: ['rev-parse', 'HEAD'] })
            .reply(200, ce)

        const { connection, cleanup } = await util.createCleanPostgresDatabase()

        try {
            const dumpManager = new DumpManager(connection)
            await dumpManager.updateCommits(
                repositoryId,
                new Map<string, Set<string>>([
                    [ca, new Set<string>()],
                    [cb, new Set<string>([ca])],
                    [cc, new Set<string>([cb])],
                    [cd, new Set<string>([cc])],
                    [ce, new Set<string>([cd])],
                ])
            )
            await util.insertDump(connection, dumpManager, repositoryId, ca, 'foo', 'test')
            await util.insertDump(connection, dumpManager, repositoryId, cb, 'foo', 'test')
            await util.insertDump(connection, dumpManager, repositoryId, cc, 'bar', 'test')

            const tipCommit = await dumpManager.discoverTip({
                repositoryId,
                frontendUrl: 'frontend',
            })
            if (!tipCommit) {
                throw new Error('Expected a tip commit')
            }
            await dumpManager.updateDumpsVisibleFromTip(repositoryId, tipCommit)

            const d1 = await dumpManager.getDump(repositoryId, ca, 'foo/test.ts')
            const d2 = await dumpManager.getDump(repositoryId, cb, 'foo/test.ts')
            const d3 = await dumpManager.getDump(repositoryId, cc, 'bar/test.ts')

            expect(d1?.visibleAtTip).toBeFalsy()
            expect(d2?.visibleAtTip).toBeTruthy()
            expect(d3?.visibleAtTip).toBeTruthy()
        } finally {
            await cleanup()
        }
    })
})
