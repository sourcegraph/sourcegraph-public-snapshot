import { createConnection } from './connection'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel } from './models.database'
import { Edge, Vertex } from 'lsif-protocol'
import { importLsif } from './importer'
import { Package, SymbolReferences, XrepoDatabase } from './xrepo'
import { Readable } from 'stream'
import { readline } from 'mz'

export async function convertLsif(
    input: Readable,
    database: string
): Promise<{ packages: Package[]; references: SymbolReferences[] }> {
    const connection = await createConnection(database, [
        DefinitionModel,
        DocumentModel,
        MetaModel,
        ReferenceModel,
        ResultChunkModel,
    ])

    try {
        await connection.query('PRAGMA synchronous = OFF')
        await connection.query('PRAGMA journal_mode = OFF')

        return await connection.transaction(entityManager =>
            importLsif(entityManager, parseLines(readline.createInterface({ input })))
        )
    } finally {
        await connection.close()
    }
}

export async function addToXrepoDatabase(
    xrepoDatabase: XrepoDatabase,
    packages: Package[],
    references: SymbolReferences[],
    repository: string,
    commit: string
): Promise<void> {
    // These need to be done in sequence as multiple write transactions
    // tends to mess up the sqlite database.

    await xrepoDatabase.addPackages(repository, commit, packages)
    await xrepoDatabase.addReferences(repository, commit, references)
}

/**
 * Converts streaming JSON input into an iterable of vertex and edge objects.
 *
 * @param lines The stream of raw, uncompressed JSON lines.
 */
async function* parseLines(lines: AsyncIterable<string>): AsyncIterable<Vertex | Edge> {
    let i = 0
    for await (const line of lines) {
        try {
            yield JSON.parse(line)
        } catch (e) {
            throw Object.assign(
                new Error(`Failed to process line #${i + 1} (${JSON.stringify(line)}): Invalid JSON.`),
                { status: 422 }
            )
        }

        i++
    }
}
