// Reads the schema.graphql file and generates the TypeScript types and a JSON file for IntelliSense

import { generateNamespace } from '@gql2ts/from-schema'
import { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } from '@gql2ts/language-typescript'
import { watch } from 'chokidar'
import { buildSchema, graphql, introspectionQuery, IntrospectionQuery } from 'graphql'
import { debounce } from 'lodash'
import { readFile, writeFile } from 'mz/fs'
import { format, resolveConfig } from 'prettier'
Object.assign(global, require('abort-controller'))

const onTerminatingSignal = (handler: () => void): void => {
    for (const signal of ['SIGTERM', 'SIGINT', 'SIGHUP'] as NodeJS.Signals[]) {
        process.on(signal, handler)
    }
}

const GRAPHQL_SCHEMA_PATH = __dirname + '/../cmd/frontend/internal/graphqlbackend/schema.graphql'

async function main(): Promise<void> {
    let abortController = new AbortController()
    onTerminatingSignal(() => abortController.abort())
    await generate(abortController.signal)
    if (process.argv.includes('--watch')) {
        const watcher = watch(GRAPHQL_SCHEMA_PATH)
        onTerminatingSignal(() => watcher.close())
        watcher.on(
            'change',
            debounce(
                () => {
                    abortController.abort()
                    abortController = new AbortController()
                    generate(abortController.signal).catch(err => console.error(err))
                },
                200,
                { leading: true }
            )
        )
    }
}

async function generate(signal: AbortSignal): Promise<void> {
    const schemaStr = await readFile(GRAPHQL_SCHEMA_PATH, 'utf8')
    const schema = buildSchema(schemaStr)
    const result = (await graphql(schema, introspectionQuery)) as { data: IntrospectionQuery }
    const json = JSON.stringify(result, null, 2)

    const formatOptions = (await resolveConfig(__dirname, { config: __dirname + '/../prettier.config.js' }))!
    const typings =
        'export type ID = string\n\n' +
        generateNamespace(
            '',
            result,
            {
                typeMap: {
                    ...DEFAULT_TYPE_MAP,
                    ID: 'ID',
                },
            },
            {
                generateNamespace: (name: string, interfaces: string) => interfaces,
                interfaceBuilder: (name: string, body: string) =>
                    'export ' + DEFAULT_OPTIONS.interfaceBuilder(name, body),
                enumTypeBuilder: (name: string, values: string) =>
                    'export ' + DEFAULT_OPTIONS.enumTypeBuilder(name, values),
                typeBuilder: (name: string, body: string) => 'export ' + DEFAULT_OPTIONS.typeBuilder(name, body),
                wrapList: (type: string) => `${type}[]`,
                postProcessor: (code: string) => format(code, { ...formatOptions, parser: 'typescript' }),
            }
        )
    if (signal.aborted) {
        return
    }
    await Promise.all([
        writeFile(__dirname + '/graphqlschema.json', json),
        writeFile(__dirname + '/src/backend/graphqlschema.ts', typings),
    ])
    console.log('Updated GraphQL typings')
}

main().catch(err => {
    console.error(err)
    process.exitCode = 1
})
