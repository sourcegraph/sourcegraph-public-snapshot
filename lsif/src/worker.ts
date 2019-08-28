import exitHook from 'async-exit-hook'
import * as temp from 'temp'
import { ConnectionCache } from './cache'
import { fs, zlib, readline } from 'mz'
import { Worker } from 'node-resque'
import { makeFilename } from './util'
import { XrepoDatabase } from './xrepo'
import { Importer } from './importer'
import { DocumentModel, RefModel, MetaModel, DefModel } from './models'
import { Edge, Vertex } from 'lsif-protocol'
import { createConnection } from 'typeorm'
import { REDIS_HOST, REDIS_PORT } from './settings'

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
            const { packages, references } = await connection.transaction(async entityManager => {
                const importer = new Importer(entityManager)

                let element: Vertex | Edge
                for await (const line of readline.createInterface({ input })) {
                    try {
                        element = JSON.parse(line)
                    } catch (e) {
                        throw new Error(`Parsing failed for line:\n${line}`)
                    }

                    try {
                        await importer.insert(element)
                    } catch (e) {
                        throw new Error(`Failed to process line:\n${line}\nCaused by:\n${e}`)
                    }
                }

                return await importer.finalize()
            })

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

main().catch(e => console.error(e))
