import { ConnectionCache, DocumentCache } from './cache'
import { createBackend } from './backend'
import { lsp } from 'lsif-protocol'
import * as fs from 'mz/fs'
import * as zlib from 'mz/zlib'

describe('C++ Queries', () => {
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)

    beforeAll(async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const input = fs.createReadStream('./test-data/cpp/data/data.lsif.gz').pipe(zlib.createGunzip())
        await backend.insertDump(input, 'five', makeCommit('five'))
    })

    test('definition of `five`', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('five', makeCommit('five'))
        const definitions = await db.definitions('main.cpp', { line: 11, character: 3 })
        // TODO - (FIXME) currently the dxr indexer returns zero-width ranges
        expect(definitions).toEqual([makeLocation('five.cpp', 2, 4, 2, 4)])
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
