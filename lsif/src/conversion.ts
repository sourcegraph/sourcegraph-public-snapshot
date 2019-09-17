import { createConnection } from './connection'
import { DefinitionModel, DocumentModel, MetaModel, ReferenceModel, ResultChunkModel } from './models.database'
import { Package, SymbolReferences } from './xrepo'
import { Readable } from 'stream'
import { importLsif } from './importer'

/**
 * Populate a SQLite database with the given input stream. Returns the
 * data required to populate the cross-repo database.
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

        return await connection.transaction(entityManager => importLsif(entityManager, input))
    } finally {
        await connection.close()
    }
}
