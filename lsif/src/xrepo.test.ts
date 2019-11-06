import rmfr from 'rmfr'
import { XrepoDatabase } from './xrepo'
import { MAX_TRAVERSAL_LIMIT } from './constants'
import { createCleanPostgresDatabase, createCommit, truncatePostgresTables, createStorageRoot } from './test-utils'
import { Connection } from 'typeorm'
import { fail } from 'assert'
import { pick } from 'lodash'
import { LsifDump } from './xrepo.models'

describe('XrepoDatabase', () => {
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase

    beforeAll(async () => {
        ;({ connection, cleanup } = await createCleanPostgresDatabase())
        storageRoot = await createStorageRoot()
        xrepoDatabase = new XrepoDatabase(storageRoot, connection)
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

        // Add relations
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

        // Add dumps
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

        // Add relations
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

        // Add dumps
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
        const fields = ['repository', 'commit', 'root']

        // Add relations
        await xrepoDatabase.updateCommits('foo', [[ca, ''], [cb, ca]])

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

        // Add relations
        await xrepoDatabase.updateCommits('foo', commits)

        // Add dumps
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

        // Add closer dump
        await xrepoDatabase.insertDump('foo', c1, '')

        // Now commit 1 should be found
        expect((await xrepoDatabase.findClosestDump('foo', cmax, 'file'))!.commit).toEqual(c1)
    })

    it('should prune overlapping roots during visibility check', async () => {
        // This database has the following commit graph:
        //
        // a -- b -- c -- d -- e -- f -- g

        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')
        const cf = createCommit('f')
        const cg = createCommit('g')

        // Add relations
        await xrepoDatabase.updateCommits('foo', [[ca, ''], [cb, ca], [cc, cb], [cd, cc], [ce, cd], [cf, ce], [cg, cf]])

        // Add dumps
        await xrepoDatabase.insertDump('foo', ca, 'r1')
        await xrepoDatabase.insertDump('foo', cb, 'r2')
        await xrepoDatabase.insertDump('foo', cc, '') // overwrites r1, r2
        const d1 = await xrepoDatabase.insertDump('foo', cd, 'r3') // overwrites ''
        const d2 = await xrepoDatabase.insertDump('foo', cf, 'r4')
        await xrepoDatabase.insertDump('foo', cg, 'r5') // not traversed

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cf)
        const visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([d1.id, d2.id])
    })

    it('should traverse branching paths during visibility check', async () => {
        // This database has the following commit graph:
        //
        // a --+-- [b] --- c ---+
        //     |                |
        //     +--- d --- [e] --+ -- [h] --+-- [i]
        //     |                           |
        //     +-- [f] --- g --------------+

        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')
        const ch = createCommit('fx')
        const ci = createCommit('i')
        const cf = createCommit('f')
        const cg = createCommit('g')

        // Add relations
        await xrepoDatabase.updateCommits('foo', [
            [ca, ''],
            [cb, ca],
            [cc, cb],
            [cd, ca],
            [ce, cd],
            [ch, cc],
            [ch, ce],
            [ci, ch],
            [ci, cg],
            [cf, ca],
            [cg, cf],
        ])

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
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([dump1.id, dump2.id, dump3.id, dump4.id])
    })

    it('should not set dumps visible farther than MAX_TRAVERSAL_LIMIT', async () => {
        if (!xrepoDatabase) {
            fail('failed beforeAll')
        }

        // This repository has the following commit graph (ancestors to the left):
        //
        // (MAX_TRAVERSAL_LIMIT + 1) -- ... -- 2 -- 1 -- 0
        //
        // Note: we use '.' as a suffix for commit numbers on construction so that
        // we can distinguish `1` and `11` (`1.1.1...` and `11.11.11...`).

        const c0 = createCommit('0.')
        const c1 = createCommit('1.')
        const cpen = createCommit(`${MAX_TRAVERSAL_LIMIT - 1}.`)
        const cmax = createCommit(`${MAX_TRAVERSAL_LIMIT}.`)

        const commits: [string, string][] = Array.from({ length: MAX_TRAVERSAL_LIMIT + 1 }, (_, i) => [
            createCommit(`${i}.`),
            createCommit(`${i + 1}.`),
        ])

        // Add relations
        await xrepoDatabase.updateCommits('foo', commits)

        // Add dumps
        const dump1 = await xrepoDatabase.insertDump('foo', cmax, '')

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cmax)
        let visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c1)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([dump1.id])

        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c0)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([])

        // Add closer dump
        const dump2 = await xrepoDatabase.insertDump('foo', cpen, '')

        // Now commit cpen should be found
        await xrepoDatabase.updateDumpsVisibleFromTip('foo', c0)
        visibleDumps = await xrepoDatabase.getVisibleDumps('foo')
        expect(visibleDumps.map((dump: LsifDump) => dump.id).sort()).toEqual([dump2.id])
    })

    it('should respect bloom filter', async () => {
        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')
        const cd = createCommit('d')
        const ce = createCommit('e')
        const cf = createCommit('f')

        const updatePackages = (commit: string, root: string, identifiers: string[]): Promise<LsifDump> =>
            xrepoDatabase.addPackagesAndReferences(
                'foo',
                commit,
                root,
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

        await xrepoDatabase.updateCommits('foo', [[ca, ''], [cb, ca], [cc, cb], [cd, cc], [ce, cd], [cf, ce]])
        await xrepoDatabase.updateDumpsVisibleFromTip('foo', cf)

        // only references containing identifier y
        expect(await getReferencedDumpIds()).toEqual([dumpa.id, dumpb.id, dumpf.id])
    })

    it('should re-query if bloom filter prunes too many results', async () => {
        const updatePackages = (commit: string, root: string, identifiers: string[]): Promise<LsifDump> =>
            xrepoDatabase.addPackagesAndReferences(
                'foo',
                commit,
                root,
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

            const dump = await updatePackages(createCommit('0'), `r${i}`, ['x', isUse ? 'y' : 'z'])
            dump.visibleAtTip = true
            await connection.getRepository(LsifDump).save(dump)

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
        const ca = createCommit('a')
        const cb = createCommit('b')
        const cc = createCommit('c')

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

        const dumpa = await xrepoDatabase.addPackagesAndReferences('foo', ca, '', [], references)
        const dumpb = await xrepoDatabase.addPackagesAndReferences('foo', cb, '', [], references)
        const dumpc = await xrepoDatabase.addPackagesAndReferences('foo', cc, '', [], references)

        const getReferencedDumpIds = async () =>
            (await xrepoDatabase.getReferences({
                repository: '',
                scheme: 'npm',
                name: 'p1',
                version: '0.1.0',
                identifier: 'y',
                limit: 50,
                offset: 0,
            })).references
                .map(reference => reference.dump_id)
                .sort()

        const updateVisibility = async (visibleA: boolean, visibleB: boolean, visibleC: boolean) => {
            dumpa.visibleAtTip = visibleA
            dumpb.visibleAtTip = visibleB
            dumpc.visibleAtTip = visibleC
            await connection.getRepository(LsifDump).save(dumpa)
            await connection.getRepository(LsifDump).save(dumpb)
            await connection.getRepository(LsifDump).save(dumpc)
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
