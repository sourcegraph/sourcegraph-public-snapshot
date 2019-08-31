import { ConnectionCache, DocumentCache } from './cache'
import { createBackend } from './backend'
import { Readable } from 'stream'
import { lsp } from 'lsif-protocol'
import * as fs from 'mz/fs'
import * as zlib from 'mz/zlib'

describe('TypeScript Queries', () => {
    const connectionCache = new ConnectionCache(10)
    const documentCache = new DocumentCache(10)

    beforeAll(async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const inputs: { input: Readable; repository: string; commit: string }[] = []

        for (const repository of ['a', 'b1', 'b2', 'b3', 'c1', 'c2', 'c3']) {
            const input = fs
                .createReadStream(`./test-data/typescript/data/${repository}.lsif.gz`)
                .pipe(zlib.createGunzip())
            const commit = makeCommit(repository)
            inputs.push({ input, repository, commit })
        }

        for (const { input, repository, commit } of inputs) {
            await backend.insertDump(input, repository, commit)
        }
    })

    test('definition of `add` from repo b1', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('b1', makeCommit('b1'))
        const definitions = await db.definitions('src/index.ts', { line: 3, character: 12 })
        expect(definitions).toEqual([makeRemoteLocation('a', 'src/index.ts', 0, 16, 0, 19)])
    })

    test('definition of `mul` from repo b1', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('b1', makeCommit('b1'))
        const definitions = await db.definitions('src/index.ts', { line: 3, character: 16 })
        expect(definitions).toEqual([makeRemoteLocation('a', 'src/index.ts', 4, 16, 4, 19)])
    })

    test('references of `mul` from repo a', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('a', makeCommit('a'))
        const references = await db.references('src/index.ts', { line: 4, character: 19 })
        // TODO - why are these garbage results in the index
        const testReferences = references!.filter(l => !l.uri.includes('node_modules'))

        // TODO - should the definition be in this result set?
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 4, 16, 4, 19)) // def
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(testReferences.length).toEqual(10)
    })

    test('references of `mul` from repo b1', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('b1', makeCommit('b1'))
        const references = await db.references('src/index.ts', { line: 3, character: 16 })
        // TODO - why are these garbage results in the index
        const testReferences = references!.filter(l => !l.uri.includes('node_modules'))

        // TODO - should the definition be in this result set?
        expect(testReferences).toContainEqual(makeRemoteLocation('a', 'src/index.ts', 4, 16, 4, 19)) // def
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 26, 3, 29)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 0, 14, 0, 17)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 15, 3, 18)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 26, 3, 29)) // 2nd use

        // Ensure no additional references
        expect(testReferences.length).toEqual(10)
    })

    test('references of `add` from repo a', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('a', makeCommit('a'))
        const references = await db.references('src/index.ts', { line: 0, character: 17 })
        // TODO - why are these garbage results in the index
        const testReferences = references!.filter(l => !l.uri.includes('node_modules'))

        // TODO - should the definition be in this result set?
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 0, 16, 0, 19)) // def
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('c1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c1', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c1', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(testReferences.length).toEqual(20)
    })

    test('references of `add` from repo c1', async () => {
        const backend = await createBackend(connectionCache, documentCache)
        const db = await backend.createDatabase('c1', makeCommit('c1'))
        const references = await db.references('src/index.ts', { line: 3, character: 16 })
        // TODO - why are these garbage results in the index
        const testReferences = references!.filter(l => !l.uri.includes('node_modules'))

        // TODO - should the definition be in this result set?
        expect(testReferences).toContainEqual(makeRemoteLocation('a', 'src/index.ts', 0, 16, 0, 19)) // def
        expect(testReferences).toContainEqual(makeRemoteLocation('a', 'src/index.ts', 11, 18, 11, 21)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b1', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('b3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeLocation('src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c2', 'src/index.ts', 3, 26, 3, 29)) // 3rd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 0, 9, 0, 12)) // import
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 11, 3, 14)) // 1st use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 15, 3, 18)) // 2nd use
        expect(testReferences).toContainEqual(makeRemoteLocation('c3', 'src/index.ts', 3, 26, 3, 29)) // 3rd use

        // Ensure no additional references
        expect(testReferences.length).toEqual(20)
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

function makeRemoteLocation(
    repository: string,
    path: string,
    startLine: number,
    startCharacter: number,
    endLine: number,
    endCharacter: number
): lsp.Location {
    return makeLocation(
        `git://${repository}?${makeCommit(repository)}#${path}`,
        startLine,
        startCharacter,
        endLine,
        endCharacter
    )
}

function makeCommit(repository: string): string {
    return repository.repeat(40).substring(0, 40)
}
