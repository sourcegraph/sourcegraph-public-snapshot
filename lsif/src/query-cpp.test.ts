import * as fs from 'mz/fs'
import * as rimraf from 'rimraf'
import * as zlib from 'mz/zlib'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { createBackend } from './backend'
import { lsp } from 'lsif-protocol'

describe('Database', () => {
    let storageRoot!: string
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)
    const resultChunkCache = new ResultChunkCache(10)

    beforeAll(async () => {
        storageRoot = await fs.promises.mkdtemp('cpp-')
        const backend = await createBackend(storageRoot, connectionCache, documentCache, resultChunkCache)
        const input = fs.createReadStream('./test-data/cpp/data/data.lsif.gz').pipe(zlib.createGunzip())
        await backend.insertDump(input, 'five', createCommit('five'))
    })

    afterAll(() => {
        rimraf.sync(storageRoot)
    })

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

//
// Helpers

function createLocation(
    uri: string,
    startLine: number,
    startCharacter: number,
    endLine: number,
    endCharacter: number
): lsp.Location {
    return lsp.Location.create(uri, {
        start: { line: startLine, character: startCharacter },
        end: { line: endLine, character: endCharacter },
    })
}

function createCommit(repository: string): string {
    return repository.repeat(40).substring(0, 40)
}
