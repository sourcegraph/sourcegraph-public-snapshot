import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { convertLsif } from './importer'
import { createCommit, createLocation, getTestData, getCleanSqliteDatabase } from './test-utils'
import { createDatabaseFilename } from './util'
import { Database } from './database'
import { entities } from './xrepo.models'
import { XrepoDatabase } from './xrepo'

describe('Database', () => {
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase
    const repository = 'five'
    const commit = createCommit('five')

    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.mkdtemp('cpp-', { encoding: 'utf8' })
        xrepoDatabase = new XrepoDatabase(await getCleanSqliteDatabase(storageRoot, entities))

        const input = await getTestData('cpp/data/data.lsif.gz')
        const database = createDatabaseFilename(storageRoot, repository, commit)
        const { packages, references } = await convertLsif(input, database)
        await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
    })

    afterAll(async () => await rmfr(storageRoot))

    const loadDatabase = (repository: string, commit: string): Database =>
        new Database(
            storageRoot,
            xrepoDatabase,
            connectionCache,
            documentCache,
            resultChunkCache,
            repository,
            commit,
            createDatabaseFilename(storageRoot, repository, commit)
        )

    it('should find all defs of `four` from main.cpp', async () => {
        const db = loadDatabase(repository, commit)
        const definitions = await db.definitions('main.cpp', { line: 12, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('main.cpp', 6, 4, 6, 4)])
    })

    it('should find all defs of `five` from main.cpp', async () => {
        const db = loadDatabase(repository, commit)
        const definitions = await db.definitions('main.cpp', { line: 11, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('five.cpp', 2, 4, 2, 4)])
    })

    it('should find all refs of `five` from main.cpp', async () => {
        const db = loadDatabase(repository, commit)
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
