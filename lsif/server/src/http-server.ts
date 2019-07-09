import express from 'express'
import asyncHandler from 'express-async-handler'
import bodyParser from 'body-parser'
import cors from 'cors'
import { JsonDatabase } from './json'
import { Database } from './database'
import { fs } from 'mz'
import * as path from 'path'
import LRU from 'lru-cache'
import { withFile } from 'tmp-promise'

// TODO add docstrings

const storageRoot = 'lsif'
// The storage size can exceed this max if a single LSIF file is larger than
// this, otherwise disk usage will be kept under this.
//
// TODO make this configurable
const softMaxStorageSize = 100 * 1024 * 1024 * 1024

// TODO make this configurable
const maxFileSize = 100 * 1024 * 1024
// disk usage to memory usage is a ratio of roughly 1:3 (based on sourcegraph/codeintellify)
// TODO make this configurable
const softMaxDBSizeLoadedIntoMemory = 100 * 1024 * 1024

interface Repository {
    repository: string
}
interface Commit {
    commit: string
}
interface RepositoryCommit extends Repository, Commit {}

async function enforceMaxDiskUsage({
    flatDirectory,
    max,
    onBeforeDelete,
}: {
    flatDirectory: string
    max: number
    onBeforeDelete: (filePath: string) => void
}): Promise<string[]> {
    if (!(await fs.exists(flatDirectory))) {
        return []
    }
    const files = await Promise.all(
        (await fs.readdir(flatDirectory)).map(async f => ({
            path: path.join(flatDirectory, f),
            stat: await fs.stat(path.join(flatDirectory, f)),
        }))
    )
    let totalSize = files.reduce((subtotal, f) => subtotal + f.stat.size, 0)
    const deletedFiles = []
    for (const f of files.sort((a, b) => a.stat.mtimeMs - b.stat.mtimeMs)) {
        if (totalSize <= max) {
            break
        }
        onBeforeDelete(f.path)
        await fs.unlink(f.path)
        totalSize = totalSize - f.stat.size
        deletedFiles.push(f.path)
    }
    return deletedFiles
}

function diskKey({ repository, commit }: RepositoryCommit): string {
    const base64Repository = Buffer.from(repository).toString('base64')
    return path.join(storageRoot, `base64repository:${base64Repository},commit:${commit}.lsif`)
}

async function createDB(repositoryCommit: RepositoryCommit): Promise<Database> {
    const db = new JsonDatabase()
    await db.load(diskKey(repositoryCommit), projectRoot => ({
        toDatabase: path_ => projectRoot + '/' + path_,
        fromDatabase: path_ => (path_.startsWith(projectRoot) ? path_.slice(`${projectRoot}/`.length) : path_),
    }))
    return db
}

const supportedMethods = ['hover', 'definitions', 'references']

function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

interface LRUDBEntry {
    dbPromise: Promise<Database>
    length: number
    dispose: () => void
}

const dbLRU = new LRU<String, LRUDBEntry>({
    max: softMaxDBSizeLoadedIntoMemory,
    length: (entry, key) => entry.length,
    dispose: (key, entry) => entry.dispose(),
})

async function withDB(repositoryCommit: RepositoryCommit, action: (db: Database) => Promise<void>): Promise<void> {
    const entry = dbLRU.get(diskKey(repositoryCommit))
    if (entry) {
        await action(await entry.dbPromise)
    } else {
        //here
        const length = (await fs.stat(diskKey(repositoryCommit))).size
        const dbPromise = createDB(repositoryCommit)
        dbLRU.set(diskKey(repositoryCommit), {
            dbPromise: dbPromise,
            length: length,
            dispose: () => dbPromise.then(db => db.close()),
        })
        await action(await dbPromise)
    }
}

function main() {
    const app = express()

    app.use(bodyParser.json({ limit: '1mb' }))
    app.use(cors())

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.post(
        '/request',
        asyncHandler(async (req, res) => {
            const { repository, commit } = req.query

            const { method, params } = req.body

            checkRepository(repository)
            checkCommit(commit)
            if (!supportedMethods.includes(method)) {
                throw Object.assign(new Error('method must be one of ' + supportedMethods), { status: 400 })
            }

            try {
                await withDB({ repository, commit }, async db => {
                    let result: any
                    switch (method) {
                        case 'hover':
                            result = db.hover(params[0], params[1])
                            break
                        case 'definitions':
                            result = db.definitions(params[0], params[1])
                            break
                        case 'references':
                            result = db.references(params[0], params[1], { includeDeclaration: false })
                            break
                        default:
                            res.status(500).send({ error: `Unknown method ${method}` })
                            return
                    }
                    res.send(result || { error: 'No result found' })
                })
            } catch (e) {
                if ('code' in e && e.code === 'ENOENT') {
                    res.send({ error: `No LSIF data available for ${repository}@${commit}.` })
                    return
                }
                throw e
            }
        })
    )

    app.get(
        '/haslsif',
        asyncHandler(async (req, res) => {
            const { repository, commit } = req.query

            const file = req.body.file

            checkRepository(repository)
            checkCommit(commit)

            if (!file) {
                res.send(await fs.exists(diskKey({ repository, commit })))
                return
            }

            if (typeof file !== 'string') {
                throw Object.assign(new Error('file must be a string'), { status: 400 })
            }

            res.send(Boolean((await createDB({ repository, commit })).stat(file)))
        })
    )

    app.post(
        '/upload',
        asyncHandler(async (req, res) => {
            const { repository, commit } = req.query

            checkRepository(repository)
            checkCommit(commit)

            const contentLength = parseInt(req.header('Content-Length') || '') || 0
            if (contentLength > maxFileSize) {
                throw Object.assign(
                    new Error(
                        `The size of the given LSIF file (${contentLength} bytes) exceeds the max of ${maxFileSize}`
                    ),
                    { status: 400 }
                )
            }
            // TODO enforce max disk usage per-repository. Currently, a
            // misbehaving client could upload a bunch of LSIF files for one
            // repository and take up all of the disk space, causing all other
            // LSIF files to get deleted to make room for the new files.
            await enforceMaxDiskUsage({
                flatDirectory: storageRoot,
                max: Math.max(0, softMaxStorageSize - contentLength),
                onBeforeDelete: filePath =>
                    console.log(`Deleting ${filePath} to help keep disk usage under ${softMaxStorageSize}.`),
            })

            await withFile(async tempFile => {
                // TODO bail early if the request body contains more data than was
                // specified in the Content-Length header. Currently, if a client
                // lies about the Content-Length, they can send a huge body that
                // fills up the disk.
                const stream = req.pipe(fs.createWriteStream(tempFile.path))
                await new Promise((resolve, reject) => {
                    stream.on('close', async () => {
                        if ((await fs.stat(tempFile.path)).size > contentLength) {
                            reject(
                                Object.assign(
                                    new Error(
                                        `The size of the given LSIF file (${contentLength} bytes) exceeds the specified Content-Length ${contentLength}`
                                    ),
                                    { status: 400 }
                                )
                            )
                            return
                        }
                        resolve()
                    })
                    stream.on('error', error => {
                        reject(error)
                    })
                })

                if (!(await fs.exists(storageRoot))) {
                    await fs.mkdir(storageRoot)
                }
                await fs.rename(tempFile.path, diskKey({ repository, commit }))

                await withDB({ repository, commit }, async () => {})

                res.send('Upload successful.')
            })
        })
    )

    // TODO make this configurable
    app.listen(5000, () => {
        console.log('Listening for HTTP requests on port 5000')
    })
}

main()
