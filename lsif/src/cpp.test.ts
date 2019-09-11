import * as fs from 'mz/fs'
import * as path from 'path'
import * as rimraf from 'rimraf'
import * as zlib from 'mz/zlib'
import { convertLsif } from './conversion'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { createCommit, createLocation } from './test-utils'
import { createDatabaseFilename } from './util'
import { Database } from './database'
import { XrepoDatabase } from './xrepo'

describe('Database', () => {
    let storageRoot!: string
    const repository = 'five'
    const commit = createCommit('five')

    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    const createDatabase = (repository: string, commit: string): Database =>
        new Database(
            storageRoot,
            new XrepoDatabase(connectionCache, path.join(storageRoot, 'correlation.db')),
            connectionCache,
            documentCache,
            resultChunkCache,
            repository,
            commit,
            createDatabaseFilename(storageRoot, repository, commit)
        )

    beforeAll(async () => {
        storageRoot = await fs.promises.mkdtemp('cpp-')
        const xrepoDatabase = new XrepoDatabase(connectionCache, path.join(storageRoot, 'correlation.db'))

        const input = fs.createReadStream('./test-data/cpp/data/data.lsif.gz').pipe(zlib.createGunzip())
        const database = createDatabaseFilename(storageRoot, repository, commit)
        const { packages, references } = await convertLsif(input, database)
        await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
    })

    afterAll(() => {
        rimraf.sync(storageRoot)
    })

    it('should find all defs of `four` from main.cpp', async () => {
        const db = createDatabase('five', createCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 12, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('main.cpp', 6, 4, 6, 4)])
    })

    it('should find all defs of `five` from main.cpp', async () => {
        const db = createDatabase('five', createCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 11, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('five.cpp', 2, 4, 2, 4)])
    })

    it('should find all refs of `five` from main.cpp', async () => {
        const db = createDatabase('five', createCommit('five'))
        const references = await db.references('main.cpp', { line: 11, character: 3 })

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(createLocation('five.h', 1, 4, 1, 8))
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(references).toContainEqual(createLocation('five.cpp', 2, 4, 2, 4))
        expect(references).toContainEqual(createLocation('main.cpp', 11, 2, 11, 6))
        expect(references).toContainEqual(createLocation('main.cpp', 13, 2, 13, 6))
        expect(references).toHaveLength(4)
    })
})
