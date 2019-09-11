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
    const repository = 'test'
    const commit = createCommit('test')

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
        storageRoot = await fs.promises.mkdtemp('typescript-')
        const xrepoDatabase = new XrepoDatabase(connectionCache, path.join(storageRoot, 'correlation.db'))

        const input = fs
            .createReadStream('./test-data/typescript/linked-references/data/test.lsif.gz')
            .pipe(zlib.createGunzip())
        const database = createDatabaseFilename(storageRoot, repository, commit)
        const { packages, references } = await convertLsif(input, database)
        await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
    })

    afterAll(() => {
        rimraf.sync(storageRoot)
    })

    it('should find all refs of `foo`', async () => {
        const db = createDatabase('data', createCommit('data'))

        const positions = [
            { line: 1, character: 5 },
            { line: 5, character: 5 },
            { line: 9, character: 5 },
            { line: 13, character: 3 },
            { line: 16, character: 3 },
        ]

        for (const position of positions) {
            const references = await db.references('src/index.ts', position)
            expect(references).toContainEqual(createLocation('src/index.ts', 1, 4, 1, 7)) // abstract def in I
            expect(references).toContainEqual(createLocation('src/index.ts', 5, 4, 5, 7)) // concrete def in A
            expect(references).toContainEqual(createLocation('src/index.ts', 9, 4, 9, 7)) // concrete def in B
            expect(references).toContainEqual(createLocation('src/index.ts', 13, 2, 13, 5)) // use via I
            expect(references).toContainEqual(createLocation('src/index.ts', 16, 2, 16, 5)) // use via B

            // Ensure no additional references
            expect(references && references.length).toEqual(5)
        }
    })
})
