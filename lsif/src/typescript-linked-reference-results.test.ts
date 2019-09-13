import * as fs from 'mz/fs'
import * as zlib from 'mz/zlib'
import rmfr from 'rmfr'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { createBackend } from './backend'
import { createCommit, createLocation } from './test-utils'

describe('Database', () => {
    let storageRoot!: string
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.promises.mkdtemp('typescript-')
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const input = fs
            .createReadStream('./test-data/typescript/linked-reference-results/data/data.lsif.gz')
            .pipe(zlib.createGunzip())
        await backend.insertDump(input, 'data', createCommit('data'))
    })

    afterAll(async () => await rmfr(storageRoot))

    it('should find all refs of `foo`', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const db = await backend.createDatabase('data', createCommit('data'))

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
