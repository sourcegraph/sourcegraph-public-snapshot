import * as temp from 'temp'
import exitHook from 'async-exit-hook'
import { ConnectionCache } from './cache'
import { createConnection, EntityManager } from 'typeorm'
import { DefModel, DocumentModel, MetaModel, RefModel } from './models'
import { Edge, Vertex } from 'lsif-protocol'
import { fs, readline, zlib } from 'mz'
import { importLsif } from './importer'
import { LOG_READY, REDIS_HOST, REDIS_PORT } from './settings'
import { makeFilename } from './util'
import { Worker } from 'node-resque'
import { XrepoDatabase } from './xrepo'

/**
 * Runs the node-resque worker that converts raw LSIF dumps into SQLite databases.
 */
async function main(): Promise<void> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    const xrepoDatabase = new XrepoDatabase(new ConnectionCache(10))
    const convertJob = { perform: makeConvertJob(xrepoDatabase) }

    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, { convert: convertJob })
    worker.on('error', e => console.error(e))
    await worker.connect()
    exitHook(() => worker.end())
    worker.start().catch(e => console.error(e))

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

/**
 * TODO
 *
 * @param backend
 */
function makeConvertJob(
    xrepoDatabase: XrepoDatabase
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        const input = fs.createReadStream(filename).pipe(zlib.createGunzip())

        const outFile = temp.path()
        const connection = await createConnection({
            database: outFile,
            entities: [DefModel, DocumentModel, MetaModel, RefModel],
            type: 'sqlite',
            name: outFile,
            synchronize: true,
        })

        try {
            const { packages, references } = await connection.transaction((entityManager: EntityManager) =>
                importLsif(entityManager, parseLines(readline.createInterface({ input })))
            )

            // These needs to be done in sequence as SQLite can only have one
            // write txn at a time without causing the other one to abort with
            // a weird error.
            await xrepoDatabase.addPackages(repository, commit, packages)
            await xrepoDatabase.addReferences(repository, commit, references)

            await fs.rename(outFile, makeFilename(repository, commit))
            await fs.unlink(filename)
        } catch (e) {
            await fs.unlink(outFile)
            throw e
        } finally {
            await connection.close()
        }
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
            yield JSON.parse(line) as Vertex | Edge
        } catch (e) {
            throw new Error(`Parsing failed for line:\n${i}`)
        }

        i++
    }
}

main().catch(e => console.error(e))
