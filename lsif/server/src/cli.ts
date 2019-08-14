import 'source-map-support/register'

import { DgraphClient, DgraphClientStub } from 'dgraph-js'
import grpc from 'grpc'
import { Edge, Vertex } from 'lsif-protocol'
import { readFile } from 'mz/fs'
import * as util from 'util'
import { getJsonSchemas } from './schema'
import { setDGraphSchema, storeLSIF } from './store'

// Loads an LSIF dump into DGraph from the terminal

async function main(): Promise<void> {
    util.inspect.defaultOptions.depth = 4
    util.inspect.defaultOptions.maxArrayLength = 20

    const filePath = process.argv[2]
    if (process.argv.length !== 5) {
        console.error('Usage: node out/cli.js <path to lsif file> <repo name> <commit id>')
        process.exitCode = 1
        return
    }

    const clientStub = new DgraphClientStub(
        // addr: optional, default: "localhost:9080"
        'localhost:9080',
        // credentials: optional, default: grpc.credentials.createInsecure()
        grpc.credentials.createInsecure()
    )
    const dgraphClient = new DgraphClient(clientStub)
    // dgraphClient.setDebugMode(true)

    await setDGraphSchema(dgraphClient)

    const contents = await readFile(filePath, 'utf-8')
    const lines = contents.trim().split('\n')
    const items = lines.map((line, index): Edge | Vertex => {
        try {
            return JSON.parse(line)
        } catch (err) {
            err.line = index + 1
            throw err
        }
    })
    console.log('items:', items.length)
    console.log('starting processing')
    const schemas = await getJsonSchemas()
    await storeLSIF({
        lsifElements: items,
        dgraphClient,
        repository: process.argv[3],
        commit: process.argv[4],
        schemas,
    })
}

// tslint:disable-next-line: no-floating-promises
main().catch(err => {
    process.exitCode = 1
    console.error(err)
})
