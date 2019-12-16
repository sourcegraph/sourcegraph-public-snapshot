import * as util from '../integration-test-util'
import * as xrepoModels from '../../../shared/models/xrepo'
import rmfr from 'rmfr'
import { Connection } from 'typeorm'
import { fail } from 'assert'
import { MAX_TRAVERSAL_LIMIT } from '../../../shared/constants'
import { pick } from 'lodash'
import { XrepoDatabase } from '../../../shared/xrepo/xrepo'

describe('XrepoDatabase', () => {
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase

    beforeAll(async () => {
        ;({ connection, cleanup } = await util.createCleanPostgresDatabase())
        storageRoot = await util.createStorageRoot()
        xrepoDatabase = new XrepoDatabase(connection, storageRoot)
    })

    afterAll(async () => {
        await rmfr(storageRoot)

        if (cleanup) {
            await cleanup()
        }
    })

    beforeEach(async () => {
        if (connection) {
            await util.truncatePostgresTables(connection)
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

        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()
        const ch = util.createCommit()

        // Add relations
        await xrepoDatabase.updateCommits(
            'foo',
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
        await xrepoDatabase.insertDump('foo', ca, '')
        await xrepoDatabase.insertDump('foo', cc, '')
        await xrepoDatabase.insertDump('foo', cg, '')

        const d1 = await xrepoDatabase.findClosestDump('foo', ca, 'file')
        const d2 = await xrepoDatabase.findClosestDump('foo', cb, 'file')
        const d3 = await xrepoDatabase.findClosestDump('foo', cc, 'file')
        const d4 = await xrepoDatabase.findClosestDump('foo', cd, 'file')
        const d5 = await xrepoDatabase.findClosestDump('foo', cf, 'file')
        const d6 = await xrepoDatabase.findClosestDump('foo', cg, 'file')
        const d7 = await xrepoDatabase.findClosestDump('foo', ce, 'file')
        const d8 = await xrepoDatabase.findClosestDump('foo', ch, 'file')

        // Test closest commit
        expect(d1?.commit).toEqual(ca)
        expect(d2?.commit).toEqual(ca)
        expect(d3?.commit).toEqual(cc)
        expect(d4?.commit).toEqual(cc)
        expect(d5?.commit).toEqual(cg)
        expect(d6?.commit).toEqual(cg)

        // Multiple nearest are chosen arbitrarily
        expect([ca, cc, cg]).toContain(d7?.commit)
        expect([ca, cc]).toContain(d8?.commit)
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

        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()
        const ch = util.createCommit()

        // Add relations
        await xrepoDatabase.updateCommits(
            'foo',
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
        await xrepoDatabase.insertDump('foo', cb, '')

        const d1 = await xrepoDatabase.findClosestDump('foo', ca, 'file')
        const d2 = await xrepoDatabase.findClosestDump('foo', cb, 'file')
        const d3 = await xrepoDatabase.findClosestDump('foo', cc, 'file')

        // Test closest commit
        expect(d1?.commit).toEqual(cb)
        expect(d2?.commit).toEqual(cb)
        expect(d3?.commit).toEqual(cb)
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

        const ca = util.createCommit()
        const cb = util.createCommit()
        const fields = ['repository', 'commit', 'root']

        // Add relations
        await xrepoDatabase.updateCommits(
            'foo',
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
            ])
        )

        // Add dumps
        await xrepoDatabase.insertDump('foo', cb, 'root1/')
        await xrepoDatabase.insertDump('foo', cb, 'root2/')

        // Test closest commit
        expect(await xrepoDatabase.findClosestDump('foo', ca, 'blah')).toBeUndefined()
        expect(pick(await xrepoDatabase.findClosestDump('foo', cb, 'root1/file.ts'), ...fields)).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root1/',
        })
        expect(pick(await xrepoDatabase.findClosestDump('foo', cb, 'root2/file.ts'), ...fields)).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root2/',
        })
        expect(pick(await xrepoDatabase.findClosestDump('foo', ca, 'root2/file.ts'), ...fields)).toEqual({
            repository: 'foo',
            commit: cb,
            root: 'root2/',
        })

        expect(await xrepoDatabase.findClosestDump('foo', ca, 'root3/file.ts')).toBeUndefined()

        await xrepoDatabase.insertDump('foo', cb, '')
        expect(pick(await xrepoDatabase.findClosestDump('foo', ca, 'root2/file.ts'), ...fields)).toEqual({
            repository: 'foo',
            commit: cb,
            root: '',
        })
        expect(pick(await xrepoDatabase.findClosestDump('foo', ca, 'root3/file.ts'), ...fields)).toEqual({
            repository: 'foo',
            commit: cb,
            root: '',
        })
    })

    it('should not return elements farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the left):
        //
        // MAX_TRAVERSAL_LIMIT -- ... -- 2 -- 1 -- 0
        //
        // Note: we use 'a' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1a1a1a...` and `11a11a11a..`).

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
        await xrepoDatabase.updateCommits('foo', commits)

        // Add dumps
        await xrepoDatabase.insertDump('foo', c0, '')

        const d1 = await xrepoDatabase.findClosestDump('foo', c0, 'file')
        const d2 = await xrepoDatabase.findClosestDump('foo', c1, 'file')
        const d3 = await xrepoDatabase.findClosestDump('foo', cpen, 'file')

        // Test closest commit
        expect(d1?.commit).toEqual(c0)
        expect(d2?.commit).toEqual(c0)
        expect(d3?.commit).toEqual(c0)

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

        // Add closer dump
        await xrepoDatabase.insertDump('foo', c1, '')

        // Now commit 1 should be found
        const dump = await xrepoDatabase.findClosestDump('foo', cmax, 'file')
        expect(dump?.commit).toEqual(c1)
    })

    it('should prune overlapping roots during visibility check', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a -- b -- c -- d -- e -- f -- g

        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()
        const cg = util.createCommit()

        // Add relations
        await xrepoDatabase.updateCommits(
            'foo',
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
        await xrepoDatabase.insertDump('foo', ca, 'r1')
        await xrepoDatabase.insertDump('foo', cb, 'r2')
        await xrepoDatabase.insertDump('foo', cc, '') // overwrites r1, r2
        const d1 = await xrepoDatabase.insertDump('foo', cd, 'r3') // overwrites ''
        const d2 = await xrepoDatabase.insertDump('foo', cf, 'r4')
        await xrepoDatabase.insertDump('foo', cg, 'r5') // not traversed

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cf)
        const visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([d1.id, d2.id])
    })

    it('should traverse branching paths during visibility check', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This database has the following commit graph:
        //
        // a --+-- [b] --- c ---+
        //     |                |
        //     +--- d --- [e] --+ -- [h] --+-- [i]
        //     |                           |
        //     +-- [f] --- g --------------+

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
        await xrepoDatabase.updateCommits(
            'foo',
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
        await xrepoDatabase.insertDump('foo', cb, 'r2')
        const dump1 = await xrepoDatabase.insertDump('foo', ce, 'r2/a') // overwrites r2 in commit b
        const dump2 = await xrepoDatabase.insertDump('foo', ce, 'r2/b')
        await xrepoDatabase.insertDump('foo', cf, 'r1/a')
        await xrepoDatabase.insertDump('foo', cf, 'r1/b')
        const dump3 = await xrepoDatabase.insertDump('foo', ch, 'r1') // overwrites r1/{a,b} in commit f
        const dump4 = await xrepoDatabase.insertDump('foo', ci, 'r3')

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', ci)
        const visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([
            dump1.id,
            dump2.id,
            dump3.id,
            dump4.id,
        ])
    })

    it('should not set dumps visible farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the left):
        //
        // (MAX_TRAVERSAL_LIMIT + 1) -- ... -- 2 -- 1 -- 0
        //
        // Note: we use 'a' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1a1a1a...` and `11a11a11a...`).

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
        await xrepoDatabase.updateCommits('foo', commits)

        // Add dumps
        const dump1 = await xrepoDatabase.insertDump('foo', cmax, '')

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cmax)
        let visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c1)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c0)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([])

        // Add closer dump
        const dump2 = await xrepoDatabase.insertDump('foo', cpen, '')

        // Now commit cpen should be found
        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c0)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: xrepoModels.LsifDump) => dump.id).sort()).toEqual([dump2.id])
    })

    it('should respect bloom filter', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()
        const cd = util.createCommit()
        const ce = util.createCommit()
        const cf = util.createCommit()

        const updatePackages = (commit: string, root: string, identifiers: string[]): Promise<xrepoModels.LsifDump> =>
            xrepoDatabase.addPackagesAndReferences(
                'foo',
                commit,
                root,
                new Date(),
                [],
                [
                    {
                        package: {
                            scheme: 'npm',
                            name: 'p1',
                            version: '0.1.0',
                        },
                        identifiers,
                    },
                ]
            )

        // Note: roots must be unique so dumps are visible
        const dumpa = await updatePackages(ca, 'r1', ['x', 'y', 'z'])
        const dumpb = await updatePackages(cb, 'r2', ['y', 'z'])
        const dumpf = await updatePackages(cf, 'r3', ['y', 'z'])
        await updatePackages(cc, 'r4', ['x', 'z'])
        await updatePackages(cd, 'r5', ['x'])
        await updatePackages(ce, 'r6', ['x', 'z'])

        const getReferencedDumpIds = async () => {
            const { references } = await xrepoDatabase.getReferences({
                repository: '',
                scheme: 'npm',
                name: 'p1',
                version: '0.1.0',
                identifier: 'y',
                limit: 50,
                offset: 0,
            })

            return references.map(reference => reference.dump_id).sort()
        }

        await xrepoDatabase.updateCommits(
            'foo',
            new Map<string, Set<string>>([
                [ca, new Set()],
                [cb, new Set([ca])],
                [cc, new Set([cb])],
                [cd, new Set([cc])],
                [ce, new Set([cd])],
                [cf, new Set([ce])],
            ])
        )
        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cf)

        // only references containing identifier y
        expect(await getReferencedDumpIds()).toEqual([dumpa.id, dumpb.id, dumpf.id])
    })

    it('should re-query if bloom filter prunes too many results', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        const updatePackages = (commit: string, root: string, identifiers: string[]): Promise<xrepoModels.LsifDump> =>
            xrepoDatabase.addPackagesAndReferences(
                'foo',
                commit,
                root,
                new Date(),
                [],
                [
                    {
                        package: {
                            scheme: 'npm',
                            name: 'p1',
                            version: '0.1.0',
                        },
                        identifiers,
                    },
                ]
            )

        const dumps = []
        for (let i = 0; i < 250; i++) {
            // Spread out uses of `y` so that we pull back a series of pages that are
            // empty and half-empty after being filtered by the bloom filter. We will
            // have to empty pages (i < 100) followed by three pages where very third
            // uses the identifier. In all, there are fifty uses spread over 5 pages.
            const isUse = i >= 100 && i % 3 === 0

            const dump = await updatePackages(util.createCommit(), `r${i}`, ['x', isUse ? 'y' : 'z'])
            dump.visibleAtTip = true
            await connection.getRepository(xrepoModels.LsifDump).save(dump)

            if (isUse) {
                // Save use ids
                dumps.push(dump.id)
            }
        }

        const { references } = await xrepoDatabase.getReferences({
            repository: 'bar',
            scheme: 'npm',
            name: 'p1',
            version: '0.1.0',
            identifier: 'y',
            limit: 50,
            offset: 0,
        })

        expect(references.map(reference => reference.dump_id).sort()).toEqual(dumps)
    })

    it('references only returned if dumps visible at tip', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        const ca = util.createCommit()
        const cb = util.createCommit()
        const cc = util.createCommit()

        const references = [
            {
                package: {
                    scheme: 'npm',
                    name: 'p1',
                    version: '0.1.0',
                },
                identifiers: ['x', 'y', 'z'],
            },
        ]

        const dumpa = await xrepoDatabase.addPackagesAndReferences('foo', ca, '', new Date(), [], references)
        const dumpb = await xrepoDatabase.addPackagesAndReferences('foo', cb, '', new Date(), [], references)
        const dumpc = await xrepoDatabase.addPackagesAndReferences('foo', cc, '', new Date(), [], references)

        const getReferencedDumpIds = async () =>
            (
                await xrepoDatabase.getReferences({
                    repository: '',
                    scheme: 'npm',
                    name: 'p1',
                    version: '0.1.0',
                    identifier: 'y',
                    limit: 50,
                    offset: 0,
                })
            ).references
                .map(reference => reference.dump_id)
                .sort()

        const updateVisibility = async (visibleA: boolean, visibleB: boolean, visibleC: boolean) => {
            dumpa.visibleAtTip = visibleA
            dumpb.visibleAtTip = visibleB
            dumpc.visibleAtTip = visibleC
            await connection.getRepository(xrepoModels.LsifDump).save(dumpa)
            await connection.getRepository(xrepoModels.LsifDump).save(dumpb)
            await connection.getRepository(xrepoModels.LsifDump).save(dumpc)
        }

        // Set a, b visible from tip
        await updateVisibility(true, true, false)
        expect(await getReferencedDumpIds()).toEqual([dumpa.id, dumpb.id])

        // Clear a, b visible from tip, set c visible fro
        await updateVisibility(false, false, true)
        expect(await getReferencedDumpIds()).toEqual([dumpc.id])

        // Clear all visible from tip
        await updateVisibility(false, false, false)
        expect(await getReferencedDumpIds()).toEqual([])
    })
})
