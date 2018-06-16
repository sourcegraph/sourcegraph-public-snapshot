import { generateNamespace } from '@gql2ts/from-schema'
import { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } from '@gql2ts/language-typescript'
import log from 'fancy-log'
import globby from 'globby'
import { buildSchema, graphql, introspectionQuery, IntrospectionQuery } from 'graphql'
import * as gulp from 'gulp'
import httpProxyMiddleware from 'http-proxy-middleware'
import { compileFromFile } from 'json-schema-to-typescript'
// @ts-ignore
import convert from 'koa-connect'
import mkdirp from 'mkdirp-promise'
import { readFile, writeFile } from 'mz/fs'
import { stat } from 'mz/fs'
import * as path from 'path'
import PluginError from 'plugin-error'
import { format, resolveConfig } from 'prettier'
// ironically, has no published typings (but will soon)
// @ts-ignore
import tsUnusedExports from 'ts-unused-exports'
import createWebpackCompiler, { Stats } from 'webpack'
import serve from 'webpack-serve'
import webpackConfig from './webpack.config'

export const build = gulp.series(gulp.parallel(schemaTypes, graphQLTypes), webpack)
export const watch = gulp.parallel(watchSchemaTypes, watchGraphQLTypes, watchWebpack)

const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    warningsFilter: warning =>
        // This is intended, so ignore warning
        /node_modules\/monaco-editor\/.*\/editorSimpleWorker.js.*\n.*dependency is an expression/.test(warning),
    colors: true,
} as Stats.ToStringOptions
const logWebpackStats = (stats: Stats) => log(stats.toString(WEBPACK_STATS_OPTIONS))

export async function webpack(): Promise<void> {
    const compiler = createWebpackCompiler(webpackConfig)
    const stats = await new Promise<Stats>((resolve, reject) => {
        compiler.run((err, stats) => (err ? reject(err) : resolve(stats)))
    })
    logWebpackStats(stats)
    if (stats.hasErrors()) {
        throw Object.assign(new Error('Failed to compile'), { showStack: false })
    }
}

export async function watchWebpack(): Promise<void> {
    if (process.env.WEBPACK_SERVE) {
        await webpackServe()
        return
    }
    const compiler = createWebpackCompiler(webpackConfig)
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

export async function webpackServe(): Promise<void> {
    await serve({
        config: {
            ...webpackConfig,
            serve: {
                clipboard: false,
                content: '../ui/assets',
                port: 3080,
                hot: false,
                dev: {
                    publicPath: '/.assets/',
                    stats: WEBPACK_STATS_OPTIONS,
                },
                add: (app, middleware, options) => {
                    // Since we're manipulating the order of middleware added, we need to handle
                    // adding these two internal middleware functions.
                    middleware.webpack()
                    middleware.content()

                    // Proxy *must* be the last middleware added.
                    app.use(
                        convert(
                            // Proxy all requests (that are not for webpack-built assets) to the Sourcegraph
                            // frontend server, and we make the Sourcegraph appURL equal to the URL of
                            // webpack-serve. This is how webpack-serve needs to work (because it does a bit
                            // more magic in injecting scripts that use WebSockets into proxied requests).
                            httpProxyMiddleware({ target: 'http://localhost:3081', ws: true })
                        )
                    )
                },
                compiler: createWebpackCompiler(webpackConfig),
            },
        },
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

export async function unusedExports(): Promise<void> {
    // TODO(sqs): Improve our usage of ts-unused-exports when its API improves (see
    // https://github.com/pzavolinsky/ts-unused-exports/pull/17 for one possible improvement).
    const analysis: { [file: string]: string[] } = tsUnusedExports(
        path.join(__dirname, 'tsconfig.json'),
        await globby('src/**/*.{ts?(x),js?(x),json}') // paths are relative to tsconfig.json
    )
    const filesWithUnusedExports = Object.keys(analysis).sort()
    if (filesWithUnusedExports.length > 0) {
        // Convert to absolute file paths with extensions to enable clickable file paths in VS Code console
        const filesWithExtensions = await Promise.all(
            filesWithUnusedExports.map(async file => {
                for (const ext of ['ts', 'tsx']) {
                    try {
                        const fullPath = path.resolve(__dirname, `${file}.${ext}`)
                        await stat(fullPath)
                        return fullPath
                    } catch (err) {
                        continue
                    }
                }
                return file
            })
        )
        throw new PluginError(
            'ts-unused-exports',
            [
                'Unused exports found (must unexport or remove):',
                ...filesWithExtensions.map((f, i) => `${f}: ${analysis[filesWithUnusedExports[i]].join(' ')}`),
            ].join('\n\t')
        )
    }
}
