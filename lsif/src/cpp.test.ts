import * as fs from 'mz/fs'
import * as zlib from 'mz/zlib'
import rmfr from 'rmfr'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { createBackend } from './backend'
import { createCommit, createLocation, getTestData } from './test-utils'

describe('Database', () => {
    let storageRoot!: string
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.mkdtemp('cpp-', { encoding: 'utf8' })
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const input = (await getTestData('cpp/data/data.lsif.gz')).pipe(zlib.createGunzip())
        await backend.insertDump(input, 'five', createCommit('five'))
    })

    afterAll(async () => await rmfr(storageRoot))

    it('should find all defs of `four` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const db = await backend.createDatabase('five', createCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 12, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('main.cpp', 6, 4, 6, 4)])
    })

    it('should find all defs of `five` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const db = await backend.createDatabase('five', createCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 11, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([createLocation('five.cpp', 2, 4, 2, 4)])
    })

    it('should find all refs of `five` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const db = await backend.createDatabase('five', createCommit('five'))
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
