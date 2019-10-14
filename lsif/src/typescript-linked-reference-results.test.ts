import * as path from 'path'
import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { convertLsif } from './importer'
import { createCommit, createLocation, getTestData, getCleanSqliteDatabase } from './test-utils'
import { dbFilename } from './util'
import { Database } from './database'
import { entities } from './xrepo.models'
import { XrepoDatabase } from './xrepo'

describe('Database', () => {
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase
    const repository = 'test'
    const commit = createCommit('test')

    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.promises.mkdtemp('typescript-')
        xrepoDatabase = new XrepoDatabase(await getCleanSqliteDatabase(storageRoot, entities))

        const input = await getTestData('typescript/linked-reference-results/data/data.lsif.gz')
        const tmp = path.join(storageRoot, 'tmp')
        const { packages, references } = await convertLsif(input, tmp, {})
        const dumpID = await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
        await fs.rename(tmp, dbFilename(storageRoot, dumpID, repository, commit))
    })

    afterAll(async () => await rmfr(storageRoot))

    const loadDatabase = async (repository: string, commit: string): Promise<Database> => {
        const dump = await xrepoDatabase.getDump(repository, commit)
        if (!dump) {
            throw new Error(`Unknown repository@commit ${repository}@${commit}`)
        }

        return new Database(
            storageRoot,
            xrepoDatabase,
            connectionCache,
            documentCache,
            resultChunkCache,
            dump.id,
            dbFilename(storageRoot, dump.id, dump.repository, dump.commit)
        )
    }

    it('should find all refs of `foo`', async () => {
        const db = await loadDatabase(repository, commit)

        const positions = [
            { line: 1, character: 5 },
            { line: 5, character: 5 },
            { line: 9, character: 5 },
            { line: 13, character: 3 },
            { line: 16, character: 3 },
        ]

        for (const position of positions) {
            const references = await db.references('src/index.ts', position, {})
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
