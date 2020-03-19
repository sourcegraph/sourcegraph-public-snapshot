import * as sinon from 'sinon'
import * as lsif from 'lsif-protocol'
import * as pgModels from '../../shared/models/pg'
import { Backend, sortMonikers } from './backend'
import { DependencyManager } from '../../shared/store/dependencies'
import { DumpManager } from '../../shared/store/dumps'
import { Database } from './database'
import { dbFilename } from '../../shared/paths'
import { createCleanPostgresDatabase, createStorageRoot } from '../../shared/test-util'
import { Connection } from 'typeorm'
import { OrderedLocationSet, ResolvedInternalLocation } from './location'
import rmfr from 'rmfr'
import { ReferencePaginationCursor } from './cursor'
import { range } from 'lodash'

const zeroUpload: pgModels.LsifUpload = {
    id: 0,
    repositoryId: 0,
    commit: '',
    root: '',
    indexer: '',
    filename: '',
    state: 'queued',
    uploadedAt: new Date(),
    startedAt: null,
    finishedAt: null,
    failureSummary: null,
    failureStacktrace: null,
    tracingContext: '',
    visibleAtTip: false,
}

const zeroDump: pgModels.LsifDump = {
    ...zeroUpload,
    state: 'completed',
    processedAt: new Date(),
}

const zeroDocument = {
    ranges: new Map(),
    hoverResults: new Map(),
    monikers: new Map(),
    packageInformation: new Map(),
}

const zeroPackage = {
    id: 0,
    scheme: '',
    name: '',
    version: '',
    dump: null,
    dump_id: 0,
    filter: Buffer.from(''),
}

const documentWithMonikers = {
    ...zeroDocument,
    monikers: new Map([
        [50, { kind: lsif.MonikerKind.local, scheme: 'test', identifier: 'm1' }],
        [51, { kind: lsif.MonikerKind.import, scheme: 'test', identifier: 'm2' }],
        [52, { kind: lsif.MonikerKind.import, scheme: 'test', identifier: 'm3' }],
    ]),
}

const documentWithPackageInformation = {
    ...zeroDocument,
    monikers: new Map([
        [50, { kind: lsif.MonikerKind.local, scheme: 'test', identifier: 'm1' }],
        [51, { kind: lsif.MonikerKind.import, scheme: 'test', identifier: 'm2', packageInformationId: 71 }],
        [52, { kind: lsif.MonikerKind.import, scheme: 'test', identifier: 'm3' }],
    ]),
    packageInformation: new Map([[71, { name: 'pkg2', version: '0.0.1' }]]),
}

// Dummy ranges used as a return value from getRangeByPosition.
// This range's monikers correlated to the documentWithMonikers
// and documentWithPackageInformation values defined above.
const ranges = [
    {
        startLine: 10,
        startCharacter: 20,
        endLine: 10,
        endCharacter: 25,
        monikerIds: new Set([50, 51, 52]),
    },
]

const makeRange = (i: number) => ({
    start: { line: i + 1, character: (i + 1) * 10 },
    end: { line: i + 1, character: (i + 1) * 10 + 5 },
})

const createTestDatabase = (dbs: Map<pgModels.DumpId, Database>) => (dump: pgModels.LsifDump) => {
    const db = dbs.get(dump.id)
    if (!db) {
        throw new Error(`Unexpected database construction (dumpId=${dump.id})`)
    }

    return db
}

describe('Backend', () => {
    let storageRoot!: string
    let connection!: Connection
    let cleanup!: () => Promise<void>
    let dumpManager!: DumpManager
    let dependencyManager!: DependencyManager

    beforeAll(async () => {
        storageRoot = await createStorageRoot()
        ;({ connection, cleanup } = await createCleanPostgresDatabase())
    })

    afterAll(async () => {
        if (storageRoot) {
            await rmfr(storageRoot)
        }

        if (cleanup) {
            await cleanup()
        }
    })

    beforeEach(() => {
        dumpManager = new DumpManager(connection)
        dependencyManager = new DependencyManager(connection)
    })

    describe('exists', () => {
        it('should return closest dumps with file', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))
            const database2 = new Database(2, dbFilename(storageRoot, 2))
            const database3 = new Database(3, dbFilename(storageRoot, 3))
            const database4 = new Database(4, dbFilename(storageRoot, 4))

            // Commit graph traversal
            sinon.stub(dumpManager, 'findClosestDumps').returns(
                Promise.resolve([
                    { ...zeroDump, id: 1 },
                    { ...zeroDump, id: 2 },
                    { ...zeroDump, id: 3 },
                    { ...zeroDump, id: 4 },
                ])
            )

            // Path existence check
            const spy1 = sinon.stub(database1, 'exists').returns(Promise.resolve(true))
            const spy2 = sinon.stub(database2, 'exists').returns(Promise.resolve(false))
            const spy3 = sinon.stub(database3, 'exists').returns(Promise.resolve(false))
            const spy4 = sinon.stub(database4, 'exists').returns(Promise.resolve(true))

            const dumps = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(
                    new Map([
                        [1, database1],
                        [2, database2],
                        [3, database3],
                        [4, database4],
                    ])
                )
            ).exists(42, 'deadbeef', '/foo/bar/baz.ts')

            expect(dumps).toEqual([
                { ...zeroDump, id: 1 },
                { ...zeroDump, id: 4 },
            ])
            expect(spy1.args[0][0]).toEqual('/foo/bar/baz.ts')
            expect(spy2.args[0][0]).toEqual('/foo/bar/baz.ts')
            expect(spy3.args[0][0]).toEqual('/foo/bar/baz.ts')
            expect(spy4.args[0][0]).toEqual('/foo/bar/baz.ts')
        })
    })

    describe('definitions', () => {
        it('should return definitions from database', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').returns(Promise.resolve({ ...zeroDump, id: 1 }))

            // Resolving target dumps
            sinon.stub(dumpManager, 'getDumpsByIds').returns(
                Promise.resolve(
                    new Map([
                        [1, { ...zeroDump, id: 1 }],
                        [2, { ...zeroDump, id: 2 }],
                        [3, { ...zeroDump, id: 3 }],
                        [4, { ...zeroDump, id: 4 }],
                    ])
                )
            )

            // In-database definitions
            sinon.stub(database1, 'definitions').returns(
                Promise.resolve([
                    { dumpId: 1, path: '1.ts', range: makeRange(1) },
                    { dumpId: 2, path: '2.ts', range: makeRange(2) },
                    { dumpId: 3, path: '3.ts', range: makeRange(3) },
                    { dumpId: 4, path: '4.ts', range: makeRange(4) },
                ])
            )

            const locations = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(new Map([[1, database1]]))
            ).definitions(42, 'deadbeef', '/foo/bar/baz.ts', { line: 5, character: 10 }, 1)

            expect(locations).toEqual([
                { dump: { ...zeroDump, id: 1 }, path: '1.ts', range: makeRange(1) },
                { dump: { ...zeroDump, id: 2 }, path: '2.ts', range: makeRange(2) },
                { dump: { ...zeroDump, id: 3 }, path: '3.ts', range: makeRange(3) },
                { dump: { ...zeroDump, id: 4 }, path: '4.ts', range: makeRange(4) },
            ])
        })

        it('should return definitions from local moniker search', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').returns(Promise.resolve({ ...zeroDump, id: 1 }))

            // Resolving target dumps
            sinon.stub(dumpManager, 'getDumpsByIds').returns(
                Promise.resolve(
                    new Map([
                        [1, { ...zeroDump, id: 1 }],
                        [2, { ...zeroDump, id: 2 }],
                        [3, { ...zeroDump, id: 3 }],
                        [4, { ...zeroDump, id: 4 }],
                    ])
                )
            )

            // Moniker resolution
            sinon
                .stub(database1, 'getRangeByPosition')
                .returns(Promise.resolve({ document: documentWithMonikers, ranges }))

            // Moniker search
            sinon.stub(database1, 'monikerResults').returns(
                Promise.resolve({
                    locations: [
                        { dumpId: 1, path: '1.ts', range: makeRange(1) },
                        { dumpId: 2, path: '2.ts', range: makeRange(2) },
                        { dumpId: 3, path: '3.ts', range: makeRange(3) },
                        { dumpId: 4, path: '4.ts', range: makeRange(4) },
                    ],
                    count: 4,
                })
            )

            const locations = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(new Map([[1, database1]]))
            ).definitions(42, 'deadbeef', '/foo/bar/baz.ts', { line: 5, character: 10 }, 1)

            expect(locations).toEqual([
                { dump: { ...zeroDump, id: 1 }, path: '1.ts', range: makeRange(1) },
                { dump: { ...zeroDump, id: 2 }, path: '2.ts', range: makeRange(2) },
                { dump: { ...zeroDump, id: 3 }, path: '3.ts', range: makeRange(3) },
                { dump: { ...zeroDump, id: 4 }, path: '4.ts', range: makeRange(4) },
            ])
        })

        it('should return definitions from remote moniker search', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))
            const database2 = new Database(2, dbFilename(storageRoot, 2))

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').returns(Promise.resolve({ ...zeroDump, id: 1 }))

            // Resolving target dumps
            sinon.stub(dumpManager, 'getDumpsByIds').returns(
                Promise.resolve(
                    new Map([
                        [1, { ...zeroDump, id: 1 }],
                        [2, { ...zeroDump, id: 2 }],
                        [3, { ...zeroDump, id: 3 }],
                        [4, { ...zeroDump, id: 4 }],
                    ])
                )
            )

            // Moniker resolution
            sinon
                .stub(database1, 'getRangeByPosition')
                .returns(Promise.resolve({ document: documentWithPackageInformation, ranges }))

            // Package resolution
            sinon.stub(dependencyManager, 'getPackage').returns(
                Promise.resolve({
                    id: 71,
                    scheme: 'test',
                    name: 'pkg2',
                    version: '0.0.1',
                    dump: { ...zeroDump, id: 2 },
                    dump_id: 2,
                })
            )

            // Moniker search (remote database)
            sinon.stub(database2, 'monikerResults').returns(
                Promise.resolve({
                    locations: [
                        { dumpId: 1, path: '1.ts', range: makeRange(1) },
                        { dumpId: 2, path: '2.ts', range: makeRange(2) },
                        { dumpId: 3, path: '3.ts', range: makeRange(3) },
                        { dumpId: 4, path: '4.ts', range: makeRange(4) },
                    ],
                    count: 4,
                })
            )

            const locations = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(
                    new Map([
                        [1, database1],
                        [2, database2],
                    ])
                )
            ).definitions(42, 'deadbeef', '/foo/bar/baz.ts', { line: 5, character: 10 }, 1)

            expect(locations).toEqual([
                { dump: { ...zeroDump, id: 1 }, path: '1.ts', range: makeRange(1) },
                { dump: { ...zeroDump, id: 2 }, path: '2.ts', range: makeRange(2) },
                { dump: { ...zeroDump, id: 3 }, path: '3.ts', range: makeRange(3) },
                { dump: { ...zeroDump, id: 4 }, path: '4.ts', range: makeRange(4) },
            ])
        })
    })

    describe('references', () => {
        const queryAllReferences = async (
            backend: Backend,
            repositoryId: number,
            commit: string,
            path: string,
            position: lsif.lsp.Position,
            limit: number,
            remoteDumpLimit?: number,
            dumpId?: number
        ): Promise<{ locations: ResolvedInternalLocation[]; pageSizes: number[]; numPages: number }> => {
            let locations: ResolvedInternalLocation[] = []
            const pageSizes: number[] = []
            let cursor: ReferencePaginationCursor | undefined

            while (true) {
                const result = await backend.references(
                    repositoryId,
                    commit,
                    path,
                    position,
                    { limit, cursor },
                    remoteDumpLimit,
                    dumpId
                )
                if (!result) {
                    break
                }

                locations = locations.concat(result.locations)
                pageSizes.push(result.locations.length)

                if (!result.newCursor) {
                    break
                }
                cursor = result.newCursor
            }

            return { locations, pageSizes, numPages: pageSizes.length }
        }

        const assertPagedReferences = async (
            numSameRepoDumps: number,
            numRemoteRepoDumps: number,
            locationsPerDump: number,
            pageLimit: number,
            remoteDumpLimit: number
        ): Promise<void> => {
            const numDatabases = 2 + numSameRepoDumps + numRemoteRepoDumps
            const numLocations = numDatabases * locationsPerDump

            const databases = range(0, numDatabases).map(i => new Database(i + 1, dbFilename(storageRoot, i + 1)))
            const dumps = range(0, numLocations).map(i => ({ ...zeroDump, id: i + 1 }))
            const locations = range(0, numLocations).map(i => ({
                dumpId: i + 1,
                path: `${i + 1}.ts`,
                range: makeRange(i),
            }))

            const getChunk = (index: number) =>
                locations.slice(index * locationsPerDump, (index + 1) * locationsPerDump)

            const sameRepoDumps = range(0, numSameRepoDumps).map(i => ({
                ...zeroPackage,
                dump: dumps[i + 2],
                dump_id: i + 3,
            }))

            const remoteRepoDumps = range(0, numRemoteRepoDumps).map(i => ({
                ...zeroPackage,
                dump: dumps[i + numSameRepoDumps + 2],
                dump_id: i + numSameRepoDumps + 3,
            }))

            const expectedLocations = locations.map((location, i) => ({
                dump: dumps[i],
                path: location.path,
                range: location.range,
            }))

            const dumpMap = new Map(dumps.map(dump => [dump.id, dump]))
            const databaseMap = new Map(databases.map((db, i) => [i + 1, db]))
            const definitionPackage = {
                id: 71,
                scheme: 'test',
                name: 'pkg2',
                version: '0.0.1',
                dump: dumps[1],
                dump_id: 2,
            }

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').callsFake(id => {
                if (id <= dumps.length) {
                    return Promise.resolve(dumps[id - 1])
                }

                throw new Error(`Unexpected getDumpById invocation (id=${id}`)
            })

            // Resolving target dumps
            sinon.stub(dumpManager, 'getDumpsByIds').returns(Promise.resolve(dumpMap))

            // Package resolution
            sinon.stub(dependencyManager, 'getPackage').returns(Promise.resolve(definitionPackage))

            // Same-repo package references
            const sameRepoStub = sinon
                .stub(dependencyManager, 'getSameRepoRemotePackageReferences')
                .callsFake(({ limit, offset }) =>
                    Promise.resolve({
                        packageReferences: sameRepoDumps.slice(offset, offset + limit),
                        totalCount: numSameRepoDumps,
                        newOffset: offset + limit,
                    })
                )

            // Remote repo package references
            const remoteRepoStub = sinon
                .stub(dependencyManager, 'getPackageReferences')
                .callsFake(({ limit, offset }) =>
                    Promise.resolve({
                        packageReferences: remoteRepoDumps.slice(offset, offset + limit),
                        totalCount: numRemoteRepoDumps,
                        newOffset: offset + limit,
                    })
                )

            // Moniker resolution
            sinon
                .stub(databases[0], 'getRangeByPosition')
                .returns(Promise.resolve({ document: documentWithPackageInformation, ranges }))

            // Package information resolution
            sinon.stub(databases[0], 'getDocumentByPath').returns(Promise.resolve(documentWithPackageInformation))

            // Same dump results
            const referenceStub = sinon
                .stub(databases[0], 'references')
                .returns(Promise.resolve(new OrderedLocationSet(getChunk(0), true)))

            const monikerStubs: sinon.SinonStub<
                Parameters<Database['monikerResults']>,
                ReturnType<Database['monikerResults']>
            >[] = []

            // Remote dump results
            for (let i = 1; i < numDatabases; i++) {
                monikerStubs.push(
                    sinon.stub(databases[i], 'monikerResults').callsFake((model, moniker, { skip = 0, take = 10 }) =>
                        Promise.resolve({
                            locations: getChunk(i).slice(skip, skip + take),
                            count: locationsPerDump,
                        })
                    )
                )
            }

            // Read all reference pages
            const { locations: resolvedLocations, pageSizes } = await queryAllReferences(
                new Backend('', dumpManager, dependencyManager, '', createTestDatabase(databaseMap)),
                42,
                'deadbeef',
                '/foo/bar/baz.ts',
                { line: 5, character: 10 },
                pageLimit,
                remoteDumpLimit,
                1
            )

            // Ensure we get all locations
            expect(resolvedLocations).toEqual(expectedLocations)

            // Ensure all pages (except for last) are full
            const copy = Array.from(pageSizes)
            expect(copy.pop()).toBeLessThanOrEqual(pageLimit)
            expect(copy.every(v => v === pageLimit)).toBeTruthy()

            // Ensure pagination limits are respected
            const expectedCalls = (results: number, limit: number) => Math.max(1, Math.ceil(results / limit))
            expect(sameRepoStub.callCount).toEqual(expectedCalls(numSameRepoDumps, remoteDumpLimit))
            expect(remoteRepoStub.callCount).toEqual(expectedCalls(numRemoteRepoDumps, remoteDumpLimit))
            expect(referenceStub.callCount).toEqual(expectedCalls(locationsPerDump, pageLimit))
            for (const stub of monikerStubs) {
                expect(stub.callCount).toEqual(expectedCalls(locationsPerDump, pageLimit))
            }
        }

        it('should return references in source and definition dumps', () => assertPagedReferences(0, 0, 1, 10, 5))
        it('should return references in remote dumps', () => assertPagedReferences(1, 0, 1, 10, 5))
        it('should return references in remote repositories', () => assertPagedReferences(0, 1, 1, 10, 5))
        it('should page large results sets', () => assertPagedReferences(0, 0, 25, 10, 5))
        it('should page a large number of remote dumps', () => assertPagedReferences(25, 25, 25, 10, 5))
        it('should respect small page size', () => assertPagedReferences(25, 25, 25, 1, 5))
        it('should respect large page size', () => assertPagedReferences(25, 25, 25, 1000, 5))
        it('should respect small remote dumps page size', () => assertPagedReferences(25, 25, 25, 10, 1))
        it('should respect large remote dumps page size', () => assertPagedReferences(25, 25, 25, 10, 25))
    })

    describe('hover', () => {
        it('should return hover content from database', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').returns(Promise.resolve({ ...zeroDump, id: 1 }))

            // In-database hovers
            sinon.stub(database1, 'hover').returns(
                Promise.resolve({
                    text: 'hover text',
                    range: makeRange(1),
                })
            )

            const hover = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(new Map([[1, database1]]))
            ).hover(42, 'deadbeef', '/foo/bar/baz.ts', { line: 5, character: 10 }, 1)

            expect(hover).toEqual({
                text: 'hover text',
                range: makeRange(1),
            })
        })

        it('should return hover content from unique definition', async () => {
            const database1 = new Database(1, dbFilename(storageRoot, 1))
            const database2 = new Database(2, dbFilename(storageRoot, 2))

            // Loading source dump
            sinon.stub(dumpManager, 'getDumpById').returns(Promise.resolve({ ...zeroDump, id: 1 }))

            // Resolving target dumps
            sinon.stub(dumpManager, 'getDumpsByIds').returns(Promise.resolve(new Map([[2, { ...zeroDump, id: 2 }]])))

            // In-database definitions
            sinon.stub(database1, 'definitions').returns(
                Promise.resolve([
                    {
                        dumpId: 2,
                        path: '2.ts',
                        range: makeRange(2),
                    },
                ])
            )

            // Remote-database hover
            sinon.stub(database2, 'hover').returns(
                Promise.resolve({
                    text: 'hover text',
                    range: makeRange(1),
                })
            )

            const hover = await new Backend(
                '',
                dumpManager,
                dependencyManager,
                '',
                createTestDatabase(
                    new Map([
                        [1, database1],
                        [2, database2],
                    ])
                )
            ).hover(42, 'deadbeef', '/foo/bar/baz.ts', { line: 5, character: 10 }, 1)

            expect(hover).toEqual({ text: 'hover text', range: makeRange(1) })
        })
    })
})

describe('sortMonikers', () => {
    it('should order monikers by kind', () => {
        const monikers = [
            {
                kind: lsif.MonikerKind.local,
                scheme: 'npm',
                identifier: 'foo',
            },
            {
                kind: lsif.MonikerKind.export,
                scheme: 'npm',
                identifier: 'bar',
            },
            {
                kind: lsif.MonikerKind.local,
                scheme: 'npm',
                identifier: 'baz',
            },
            {
                kind: lsif.MonikerKind.import,
                scheme: 'npm',
                identifier: 'bonk',
            },
        ]

        expect(sortMonikers(monikers)).toEqual([monikers[3], monikers[0], monikers[2], monikers[1]])
    })

    it('should remove subsumed monikers', () => {
        const monikers = [
            {
                kind: lsif.MonikerKind.local,
                scheme: 'go',
                identifier: 'foo',
            },
            {
                kind: lsif.MonikerKind.local,
                scheme: 'tsc',
                identifier: 'bar',
            },
            {
                kind: lsif.MonikerKind.local,
                scheme: 'gomod',
                identifier: 'baz',
            },
            {
                kind: lsif.MonikerKind.local,
                scheme: 'npm',
                identifier: 'baz',
            },
        ]

        expect(sortMonikers(monikers)).toEqual([monikers[2], monikers[3]])
    })

    it('should not remove subsumable (but non-subsumed) monikers', () => {
        const monikers = [
            {
                kind: lsif.MonikerKind.local,
                scheme: 'go',
                identifier: 'foo',
            },
            {
                kind: lsif.MonikerKind.local,
                scheme: 'tsc',
                identifier: 'bar',
            },
        ]

        expect(sortMonikers(monikers)).toEqual(monikers)
    })
})
