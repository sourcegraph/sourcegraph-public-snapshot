import * as fs from 'mz/fs'
import * as path from 'path'
import { lsp } from 'lsif-protocol'
import { Readable } from 'stream'

/**
 * Return a filesystem read stream for the given test file. This will cover
 * the cases where `yarn test` is ran from the root or from the lsif directory.
 *
 * @param filename The path relative to test-data directory.
 */
export async function getTestData(filename: string): Promise<Readable> {
    return fs.createReadStream(path.join((await fs.exists('lsif')) ? 'lsif' : '', 'test-data', filename))
}

export function createLocation(
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

export function createRemoteLocation(
    repository: string,
    path: string,
    startLine: number,
    startCharacter: number,
    endLine: number,
    endCharacter: number
): lsp.Location {
    const url = new URL(`git://${repository}`)
    url.search = createCommit(repository)
    url.hash = path

    return createLocation(url.href, startLine, startCharacter, endLine, endCharacter)
}

export function createCommit(repository: string): string {
    return repository.repeat(40).substring(0, 40)
}
