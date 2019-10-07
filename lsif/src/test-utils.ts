import * as fs from 'mz/fs'
import * as path from 'path'
import * as uuid from 'uuid'
import { Connection } from 'typeorm'
import { createSqliteConnection } from './connection'
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

/**
 * Create a new SQLite database connection with a randomized filename and
 * connection cache key.
 *
 * @param storageRoot The directory in which to create the database.
 * @param entities The set of expected entities present in this schema.
 */
export function getCleanSqliteDatabase(
    storageRoot: string,
    // Decorators are not possible type check
    // eslint-disable-next-line @typescript-eslint/ban-types
    entities: Function[]
): Promise<Connection> {
    return createSqliteConnection(path.join(storageRoot, `${uuid.v4()}.db`), entities)
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
