import * as sqliteModels from '../../shared/models/sqlite'
import { comparePosition, findRanges, mapRangesToInternalLocations, Database } from './database'
import * as fs from 'mz/fs'
import * as nodepath from 'path'
import * as constants from '../../shared/constants'
import { convertLsif } from '../../dump-processor/conversion/importer'
import { PathExistenceChecker } from '../../dump-processor/conversion/existence'
import rmfr from 'rmfr'
import * as uuid from 'uuid'
import { createStorageRoot } from '../../shared/test-util'

describe('Database', () => {
    let storageRoot!: string
    let database!: Database

    const makeDatabase = async (filename: string): Promise<Database> => {
        // Create a filesystem read stream for the given test file. This will cover
        // the cases where `yarn test` is run from the root or from the lsif directory.
        const root = (await fs.exists('lsif')) ? 'lsif' : ''
        const sourceFile = nodepath.join(root, 'test-data', filename)
        const databaseFile = nodepath.join(storageRoot, constants.TEMP_DIR, uuid.v4())

        await convertLsif({
            path: sourceFile,
            root: '',
            database: databaseFile,
            pathExistenceChecker: new PathExistenceChecker({
                repositoryId: 42,
                commit: 'ad3507cbeb18d1ed2b8a0f6354dea88a101197f3',
                root: '',
            }),
        })

        return new Database(1, databaseFile)
    }

    beforeAll(async () => {
        storageRoot = await createStorageRoot()
        database = await makeDatabase('lsif-go@ad3507cb.lsif.gz')
    })

    afterAll(async () => {
        if (storageRoot) {
            await rmfr(storageRoot)
        }
    })

    describe('exists', () => {
        it('should check document path', async () => {
            expect(await database.exists('cmd/lsif-go/main.go')).toEqual(true)
            expect(await database.exists('internal/index/indexer.go')).toEqual(true)
            expect(await database.exists('missing.go')).toEqual(false)
        })
    })

    describe('definitions', () => {
        it('should correlate definitions', async () => {
            // `\ts, err := indexer.Index()` -> `\t Index() (*Stats, error)`
            //                      ^^^^^           ^^^^^

            expect(await database.definitions('cmd/lsif-go/main.go', { line: 110, character: 22 })).toEqual([
                {
                    dumpId: 1,
                    path: 'internal/index/indexer.go',
                    range: { start: { line: 20, character: 1 }, end: { line: 20, character: 6 } },
                },
            ])
        })
    })

    describe('references', () => {
        it('should correlate references', async () => {
            // `func (w *Writer) EmitRange(start, end Pos) (string, error) {`
            //                   ^^^^^^^^^
            //
            // -> `\t\trangeID, err := i.w.EmitRange(lspRange(ipos, ident.Name, isQuotedPkgName))`
            //                             ^^^^^^^^^
            //
            // -> `\t\t\trangeID, err = i.w.EmitRange(lspRange(ipos, ident.Name, false))`
            //                              ^^^^^^^^^

            expect((await database.references('protocol/writer.go', { line: 85, character: 20 })).locations).toEqual([
                {
                    dumpId: 1,
                    path: 'protocol/writer.go',
                    range: { start: { line: 85, character: 17 }, end: { line: 85, character: 26 } },
                },
                {
                    dumpId: 1,
                    path: 'internal/index/indexer.go',
                    range: { start: { line: 529, character: 22 }, end: { line: 529, character: 31 } },
                },
                {
                    dumpId: 1,
                    path: 'internal/index/indexer.go',
                    range: { start: { line: 380, character: 22 }, end: { line: 380, character: 31 } },
                },
            ])
        })
    })

    describe('hover', () => {
        it('should correlate hover text', async () => {
            // `\tcontents, err := findContents(pkgs, p, f, obj)`
            //                     ^^^^^^^^^^^^

            const ticks = '```'
            const docstring = 'findContents returns contents used as hover info for given object.'
            const signature =
                'func findContents(pkgs []*Package, p *Package, f *File, obj Object) ([]MarkedString, error)'

            expect(await database.hover('internal/index/indexer.go', { line: 628, character: 20 })).toEqual({
                text: `${ticks}go\n${signature}\n${ticks}\n\n---\n\n${docstring}`,
                range: { start: { line: 628, character: 18 }, end: { line: 628, character: 30 } },
            })
        })
    })

    describe('getRangeByPosition', () => {
        it('should return correct range and document with monikers', async () => {
            // `func NewMetaData(id, root string, info ToolInfo) *MetaData {`
            //       ^^^^^^^^^^^

            const { document, ranges } = await database.getRangeByPosition('protocol/protocol.go', {
                line: 92,
                character: 10,
            })

            expect(ranges).toHaveLength(1)
            expect(ranges[0].startLine).toEqual(92)
            expect(ranges[0].startCharacter).toEqual(5)
            expect(ranges[0].endLine).toEqual(92)
            expect(ranges[0].endCharacter).toEqual(16)

            const monikers = Array.from(ranges[0].monikerIds).map(id => document?.monikers.get(id))
            expect(monikers).toHaveLength(1)
            expect(monikers[0]?.kind).toEqual('export')
            expect(monikers[0]?.scheme).toEqual('gomod')
            expect(monikers[0]?.identifier).toEqual('github.com/sourcegraph/lsif-go/protocol:NewMetaData')

            const packageInformation = document?.packageInformation.get(monikers[0]?.packageInformationId || 0)
            expect(packageInformation?.name).toEqual('github.com/sourcegraph/lsif-go')
            expect(packageInformation?.version).toEqual('v0.0.0-ad3507cbeb18')
        })
    })

    describe('monikerResults', () => {
        const edgeLocations = [
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 600, character: 1 }, end: { line: 600, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 644, character: 1 }, end: { line: 644, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 507, character: 1 }, end: { line: 507, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 553, character: 1 }, end: { line: 553, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 462, character: 1 }, end: { line: 462, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 484, character: 1 }, end: { line: 484, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 410, character: 5 }, end: { line: 410, character: 9 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 622, character: 1 }, end: { line: 622, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 440, character: 1 }, end: { line: 440, character: 5 } },
            },
            {
                dumpId: 1,
                path: 'protocol/protocol.go',
                range: { start: { line: 530, character: 1 }, end: { line: 530, character: 5 } },
            },
        ]

        it('should query definitions table', async () => {
            const { locations, count } = await database.monikerResults(
                sqliteModels.DefinitionModel,
                {
                    scheme: 'gomod',
                    identifier: 'github.com/sourcegraph/lsif-go/protocol:Edge',
                },
                {}
            )

            expect(locations).toEqual(edgeLocations)
            expect(count).toEqual(10)
        })

        it('should respect pagination', async () => {
            const { locations, count } = await database.monikerResults(
                sqliteModels.DefinitionModel,
                {
                    scheme: 'gomod',
                    identifier: 'github.com/sourcegraph/lsif-go/protocol:Edge',
                },
                { skip: 3, take: 4 }
            )

            expect(locations).toEqual(edgeLocations.slice(3, 7))
            expect(count).toEqual(10)
        })

        it('should query references table', async () => {
            const { locations, count } = await database.monikerResults(
                sqliteModels.ReferenceModel,
                {
                    scheme: 'gomod',
                    identifier: 'github.com/slimsag/godocmd:ToMarkdown',
                },
                {}
            )

            expect(locations).toEqual([
                {
                    dumpId: 1,
                    path: 'internal/index/helper.go',
                    range: { start: { line: 78, character: 6 }, end: { line: 78, character: 16 } },
                },
            ])
            expect(count).toEqual(1)
        })
    })
})

describe('findRanges', () => {
    it('should return ranges containing position', () => {
        const range1 = {
            startLine: 0,
            startCharacter: 3,
            endLine: 0,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range2 = {
            startLine: 1,
            startCharacter: 3,
            endLine: 1,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range3 = {
            startLine: 2,
            startCharacter: 3,
            endLine: 2,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range4 = {
            startLine: 3,
            startCharacter: 3,
            endLine: 3,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range5 = {
            startLine: 4,
            startCharacter: 3,
            endLine: 4,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }

        expect(findRanges([range1, range2, range3, range4, range5], { line: 0, character: 4 })).toEqual([range1])
        expect(findRanges([range1, range2, range3, range4, range5], { line: 1, character: 4 })).toEqual([range2])
        expect(findRanges([range1, range2, range3, range4, range5], { line: 2, character: 4 })).toEqual([range3])
        expect(findRanges([range1, range2, range3, range4, range5], { line: 3, character: 4 })).toEqual([range4])
        expect(findRanges([range1, range2, range3, range4, range5], { line: 4, character: 4 })).toEqual([range5])
    })

    it('should order inner-most ranges first', () => {
        const range1 = {
            startLine: 0,
            startCharacter: 3,
            endLine: 4,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range2 = {
            startLine: 1,
            startCharacter: 3,
            endLine: 3,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range3 = {
            startLine: 2,
            startCharacter: 3,
            endLine: 2,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range4 = {
            startLine: 5,
            startCharacter: 3,
            endLine: 5,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }
        const range5 = {
            startLine: 6,
            startCharacter: 3,
            endLine: 6,
            endCharacter: 5,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }

        expect(findRanges([range1, range2, range3, range4, range5], { line: 2, character: 4 })).toEqual([
            range3,
            range2,
            range1,
        ])
    })
})

describe('comparePosition', () => {
    it('should return the relative order to a range', () => {
        const range = {
            startLine: 5,
            startCharacter: 11,
            endLine: 5,
            endCharacter: 13,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        }

        expect(comparePosition(range, { line: 5, character: 11 })).toEqual(0)
        expect(comparePosition(range, { line: 5, character: 12 })).toEqual(0)
        expect(comparePosition(range, { line: 5, character: 13 })).toEqual(0)
        expect(comparePosition(range, { line: 4, character: 12 })).toEqual(+1)
        expect(comparePosition(range, { line: 5, character: 10 })).toEqual(+1)
        expect(comparePosition(range, { line: 5, character: 14 })).toEqual(-1)
        expect(comparePosition(range, { line: 6, character: 12 })).toEqual(-1)
    })
})

describe('mapRangesToInternalLocations', () => {
    it('should map ranges to locations', () => {
        const ranges = new Map<sqliteModels.RangeId, sqliteModels.RangeData>()
        ranges.set(1, {
            startLine: 1,
            startCharacter: 1,
            endLine: 1,
            endCharacter: 2,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        })
        ranges.set(2, {
            startLine: 3,
            startCharacter: 1,
            endLine: 3,
            endCharacter: 2,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        })
        ranges.set(4, {
            startLine: 2,
            startCharacter: 1,
            endLine: 2,
            endCharacter: 2,
            monikerIds: new Set<sqliteModels.MonikerId>(),
        })

        const path = 'src/position.ts'
        const locations = mapRangesToInternalLocations(42, ranges, path, new Set([1, 2, 4]))
        expect(locations).toContainEqual({
            dumpId: 42,
            path,
            range: { start: { line: 1, character: 1 }, end: { line: 1, character: 2 } },
        })
        expect(locations).toContainEqual({
            dumpId: 42,
            path,
            range: { start: { line: 3, character: 1 }, end: { line: 3, character: 2 } },
        })
        expect(locations).toContainEqual({
            dumpId: 42,
            path,
            range: { start: { line: 2, character: 1 }, end: { line: 2, character: 2 } },
        })
        expect(locations).toHaveLength(3)
    })
})
