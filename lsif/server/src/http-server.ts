import 'source-map-support/register'

import { wrap } from 'async-middleware'
import bodyParser from 'body-parser'
import { DgraphClient, DgraphClientStub } from 'dgraph-js'
import express from 'express'
import morgan from 'morgan'
import { Position } from 'vscode-languageserver-types'
import { checkExists, handlers } from './query'
import { getJsonSchemas } from './schema'
import { parseLSIFStream, setDGraphSchema, storeLSIF } from './store'

const DGRAPH_ADDRESS = process.env.DGRAPH_ADDRESS || undefined

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const PORT = (process.env.LSIF_HTTP_PORT && parseInt(process.env.LSIF_HTTP_PORT, 10)) || 3186

// addr: optional, default: "localhost:9080"
// credentials: optional, default: grpc.credentials.createInsecure()
const clientStub = new DgraphClientStub(DGRAPH_ADDRESS)
const dgraphClient = new DgraphClient(clientStub)

/**
 * Throws an error with status 400 if the repository is invalid.
 */
function checkRepository(repository: unknown): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit is invalid.
 */
function checkCommit(commit: unknown): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * Runs the HTTP server which accepts LSIF file uploads and responds to
 * hover/defs/refs requests.
 */
async function main(): Promise<void> {
    await setDGraphSchema(dgraphClient)
    const schemas = await getJsonSchemas()

    const app = express()

    app.use(morgan('dev'))

    app.get('/ping', (req, res) => {
        res.send({ pong: 'pong' })
    })

    app.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(async (req, res) => {
            const { repository, commit } = req.query
            const { path, position, method } = req.body

            checkRepository(repository)
            checkCommit(commit)
            if (!Position.is(position)) {
                throw Object.assign(new Error('Invalid position'), { status: 422 })
            }
            if (typeof method !== 'string' || !Object.prototype.hasOwnProperty.call(handlers, method)) {
                throw Object.assign(new Error('Unknown method ' + method), { status: 422 })
            }

            const result = await handlers[method as keyof typeof handlers]({
                dgraphClient,
                repository,
                commit,
                path,
                position,
            })
            res.json(result)
        })
    )

    app.post(
        '/exists',
        wrap(async (req, res) => {
            const { repository, commit, file } = req.query

            checkRepository(repository)
            checkCommit(commit)

            res.json(await checkExists({ dgraphClient, repository, commit, file }))
        })
    )

    app.post(
        '/upload',
        wrap(async (req, res) => {
            const { repository, commit } = req.query

            checkRepository(repository)
            checkCommit(commit)

            if (req.header('Content-Length') && parseInt(req.header('Content-Length')!, 10) === 0) {
                throw Object.assign(new Error('No request content'), { status: 422 })
            }

            await storeLSIF({ repository, commit, lsifElements: parseLSIFStream(req), dgraphClient, schemas })
        })
    )

    // Error handler
    // tslint:disable-next-line: no-any
    app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
        if (err && err.status) {
            const { status, headers, message, ...data } = err
            res.status(err.status).send({ message, ...data })
            return
        }
        res.status(500).send({ message: 'Unknown error' })
        console.error(err)
    })

    app.listen(PORT, () => {
        console.log(`Listening for HTTP requests on port ${PORT}`)
    })
}

// tslint:disable-next-line: no-floating-promises
main().catch(err => {
    console.error(err)
    setTimeout(() => process.exit(1), 100)
})
