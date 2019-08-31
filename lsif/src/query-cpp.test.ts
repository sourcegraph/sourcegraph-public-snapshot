import * as fs from 'mz/fs'
import * as temp from 'temp'
import * as zlib from 'mz/zlib'
import { ConnectionCache, DocumentCache } from './cache'
import { createBackend } from './backend'
import { lsp } from 'lsif-protocol'

describe('Database', () => {
    let storageRoot!: string
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)

    beforeAll(async () => {
        storageRoot = temp.mkdirSync('cpp') // eslint-disable-line no-sync
        const backend = await createBackend(storageRoot, connectionCache, documentCache)
        const input = fs.createReadStream('./test-data/cpp/data/data.lsif.gz').pipe(zlib.createGunzip())
        await backend.insertDump(input, 'five', makeCommit('five'))
    })

    it('should find all defs of `four` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache)
        const db = await backend.createDatabase('five', makeCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 12, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([makeLocation('main.cpp', 6, 4, 6, 4)])
    })

    it('should find all defs of `five` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache)
        const db = await backend.createDatabase('five', makeCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 11, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([makeLocation('five.cpp', 2, 4, 2, 4)])
    })

    it('should find all refs of `five` from main.cpp', async () => {
        const backend = await createBackend(storageRoot, connectionCache, documentCache)
        const db = await backend.createDatabase('five', makeCommit('five'))
        const references = await db.references('main.cpp', { line: 11, character: 3 })

        // TODO - should the definition be in this result set?
        expect(references).toContainEqual(makeLocation('five.h', 1, 4, 1, 8))
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(references).toContainEqual(makeLocation('five.cpp', 2, 4, 2, 4))
        expect(references).toContainEqual(makeLocation('main.cpp', 11, 2, 11, 6))
        expect(references).toContainEqual(makeLocation('main.cpp', 13, 2, 13, 6))
        expect(references && references.length).toEqual(4)
    })
})

//
// Helpers

function makeLocation(
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

function makeCommit(repository: string): string {
    return repository.repeat(40).substring(0, 40)
}
