import { createConnection } from './connection'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel } from './models.database'
import { Edge, Vertex } from 'lsif-protocol'
import { importLsif } from './importer'
import { Package, SymbolReferences } from './xrepo'
import { Readable } from 'stream'
import { readline } from 'mz'

/**
 * Populate a SQLite database with the given input stream. Returns the
 * data required to populate the correlation database.
 *
 * @param input The input stream containing JSON-encoded LSIF data.
 * @param database The filepath of the database to populate.
 */
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
