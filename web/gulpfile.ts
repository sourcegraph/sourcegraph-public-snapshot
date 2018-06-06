import { generateNamespace } from '@gql2ts/from-schema'
import { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } from '@gql2ts/language-typescript'
import log from 'fancy-log'
import { buildSchema, graphql, introspectionQuery, IntrospectionQuery } from 'graphql'
import * as gulp from 'gulp'
import { compileFromFile } from 'json-schema-to-typescript'
import mkdirp from 'mkdirp-promise'
import { readFile, writeFile } from 'mz/fs'
import { format, resolveConfig } from 'prettier'
import createWebpackCompiler, { Stats } from 'webpack'
import WebpackDevServer from 'webpack-dev-server'
import webpackConfig from './webpack.config'

export const build = gulp.series(gulp.parallel(schemaTypes, graphQLTypes), webpack)
export const watch = gulp.parallel(watchSchemaTypes, watchGraphQLTypes, watchWebpack)

const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
} as Stats.ToStringOptions
const logWebpackStats = (stats: Stats) => log(stats.toString(WEBPACK_STATS_OPTIONS))

export async function webpack(): Promise<void> {
    const compiler = createWebpackCompiler(webpackConfig)
    const stats = await new Promise<Stats>((resolve, reject) => {
        compiler.run((err, stats) => (err ? reject(err) : resolve(stats)))
    })
    logWebpackStats(stats)
}

const createWatchWebpackCompiler = () => {
    const compiler = createWebpackCompiler(webpackConfig)
    compiler.hooks.watchRun.tap('log', () => log('Starting webpack compilation'))
    return compiler
}

export async function watchWebpack(): Promise<void> {
    if (process.env.USE_WEBPACK_DEV_SERVER) {
        await webpackDevServer()
        return
    }
    const compiler = createWatchWebpackCompiler()
    compiler.hooks.watchRun.tap('log', () => log('Starting webpack compilation'))
    await new Promise<never>((resolve, reject) => {
        compiler.watch({}, (err, stats) => {
            if (err) {
                reject(err)
                return
            }
            logWebpackStats(stats)
        })
    })
}

export async function webpackDevServer(): Promise<void> {
    const compiler = createWatchWebpackCompiler()
    const server = new WebpackDevServer(compiler, { ...webpackConfig.devServer, stats: WEBPACK_STATS_OPTIONS })
    await new Promise<void>((resolve, reject) => {
        server.listen(webpackConfig.devServer.port, err => (err ? reject(err) : resolve()))
    })
}

const GRAPHQL_SCHEMA_PATH = __dirname + '/../cmd/frontend/internal/graphqlbackend/schema.graphql'

export async function watchGraphQLTypes(): Promise<void> {
    await graphQLTypes()
    await new Promise<never>((resolve, reject) => {
        gulp.watch(GRAPHQL_SCHEMA_PATH, graphQLTypes).on('error', reject)
    })
}

/** Generates the TypeScript types for the GraphQL schema */
export async function graphQLTypes(): Promise<void> {
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
    await Promise.all([
        writeFile(__dirname + '/graphqlschema.json', json),
        writeFile(__dirname + '/src/backend/graphqlschema.ts', typings),
    ])
}

/** Generates the TypeScript types for the JSON schemas */
export async function schemaTypes(): Promise<void> {
    await mkdirp(__dirname + '/src/schema')
    await Promise.all(
        ['settings', 'site'].map(async file => {
            const types = await compileFromFile(__dirname + `/../schema/${file}.schema.json`, {
                cwd: __dirname + '/../schema',
            })
            await writeFile(__dirname + `/src/schema/${file}.schema.d.ts`, types)
        })
    )
}

export async function watchSchemaTypes(): Promise<void> {
    await schemaTypes()
    await new Promise<never>((resolve, reject) => {
        gulp.watch(__dirname + '/../schema/*.schema.json', schemaTypes).on('error', reject)
    })
}
