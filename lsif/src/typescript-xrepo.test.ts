import * as fs from 'mz/fs'
import rmfr from 'rmfr'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { convertLsif } from './importer'
import { createCommit, createLocation, createRemoteLocation, getTestData, getCleanSqliteDatabase } from './test-utils'
import { createDatabaseFilename } from './util'
import { Database } from './database'
import { Readable } from 'stream'
import { XrepoDatabase } from './xrepo'
import { PackageModel, ReferenceModel } from './models.xrepo'

describe('Database', () => {
    let storageRoot!: string
    let xrepoDatabase!: XrepoDatabase
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.mkdtemp('typescript-', { encoding: 'utf8' })
        xrepoDatabase = new XrepoDatabase(await getCleanSqliteDatabase(storageRoot, [PackageModel, ReferenceModel]))

        for (const { input, repository, commit } of await createTestInputs()) {
            const database = createDatabaseFilename(storageRoot, repository, commit)
            const { packages, references } = await convertLsif(input, database)
            await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
        }
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

    it('should find all defs of `add` from repo a', async () => {
        const db = loadDatabase('a', createCommit('a'))
        const definitions = await db.definitions('src/index.ts', { line: 11, character: 18 })
        expect(definitions).toContainEqual(createLocation('src/index.ts', 0, 16, 0, 19))
        expect(definitions && definitions.length).toEqual(1)
    })

    it('should find all defs of `add` from repo b1', async () => {
        const db = loadDatabase('b1', createCommit('b1'))
        const definitions = await db.definitions('src/index.ts', { line: 3, character: 12 })
        expect(definitions).toContainEqual(createRemoteLocation('a', 'src/index.ts', 0, 16, 0, 19))
        expect(definitions && definitions.length).toEqual(1)
    })

    it('should find all defs of `mul` from repo b1', async () => {
        const db = loadDatabase('b1', createCommit('b1'))
        const definitions = await db.definitions('src/index.ts', { line: 3, character: 16 })
        expect(definitions).toContainEqual(createRemoteLocation('a', 'src/index.ts', 4, 16, 4, 19))
        expect(definitions && definitions.length).toEqual(1)
    })

    it('should find all refs of `mul` from repo a', async () => {
        const db = loadDatabase('a', createCommit('a'))
        // TODO - (FIXME) why are these garbage results in the index
        const references = (await db.references('src/index.ts', { line: 4, character: 19 }))!.filter(
            l => !l.uri.includes('node_modules')
        )

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(createLocation('src/index.ts', 4, 16, 4, 19)) // def
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(references && references.length).toEqual(10)
    })

    it('should find all refs of `mul` from repo b1', async () => {
        const db = loadDatabase('b1', createCommit('b1'))
        // TODO - (FIXME) why are these garbage results in the index
        const references = (await db.references('src/index.ts', { line: 3, character: 16 }))!.filter(
            l => !l.uri.includes('node_modules')
        )

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(createRemoteLocation('a', 'src/index.ts', 4, 16, 4, 19)) // def
        expect(references).toContainEqual(createLocation('src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createLocation('src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createLocation('src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(references && references.length).toEqual(10)
    })

    it('should find all refs of `add` from repo a', async () => {
        const db = loadDatabase('a', createCommit('a'))
        // TODO - (FIXME) why are these garbage results in the index
        const references = (await db.references('src/index.ts', { line: 0, character: 17 }))!.filter(
            l => !l.uri.includes('node_modules')
        )

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(createLocation('src/index.ts', 0, 16, 0, 19)) // def
        expect(references).toContainEqual(createLocation('src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('c1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c1', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('c1', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(references && references.length).toEqual(20)
    })

    it('should find all refs of `add` from repo c1', async () => {
        const db = loadDatabase('c1', createCommit('c1'))
        // TODO - (FIXME) why are these garbage results in the index
        const references = (await db.references('src/index.ts', { line: 3, character: 16 }))!.filter(
            l => !l.uri.includes('node_modules')
        )

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(createRemoteLocation('a', 'src/index.ts', 0, 16, 0, 19)) // def
        expect(references).toContainEqual(createRemoteLocation('a', 'src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('b3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createLocation('src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createLocation('src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createLocation('src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createLocation('src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('c2', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(references).toContainEqual(createRemoteLocation('c3', 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(references && references.length).toEqual(20)
    })
})

async function createTestInputs(): Promise<
    {
        input: Readable
        repository: string
        commit: string
    }[]
> {
    const repositories = ['a', 'b1', 'b2', 'b3', 'c1', 'c2', 'c3']

    const inputs = []
    for (const repository of repositories) {
        const input = await getTestData(`typescript/xrepo/data/${repository}.lsif.gz`)
        const commit = createCommit(repository)
        inputs.push({ input, repository, commit })
    }

    return inputs
}
