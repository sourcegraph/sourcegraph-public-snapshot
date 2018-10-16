import { generateNamespace } from '@gql2ts/from-schema'
import { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } from '@gql2ts/language-typescript'
import { ChildProcess, spawn } from 'child_process'
import execa from 'execa'
import log from 'fancy-log'
import globby from 'globby'
import { buildSchema, graphql, introspectionQuery, IntrospectionQuery } from 'graphql'
import gulp from 'gulp'
import httpProxyMiddleware from 'http-proxy-middleware'
import { compile as compileJSONSchema } from 'json-schema-to-typescript'
// @ts-ignore
import convert from 'koa-connect'
import latestVersion from 'latest-version'
import mkdirp from 'mkdirp-promise'
import { readFile, stat, writeFile } from 'mz/fs'
import * as path from 'path'
import PluginError from 'plugin-error'
import { format, resolveConfig } from 'prettier'
import * as semver from 'semver'
// ironically, has no published typings (but will soon)
// @ts-ignore
import tsUnusedExports from 'ts-unused-exports'
import createWebpackCompiler, { Stats } from 'webpack'
import serve from 'webpack-serve'
import webpackConfig from './webpack.config'

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

export async function webpackServe(): Promise<void> {
    await serve(
        {},
        {
            config: {
                ...webpackConfig,
                serve: {
                    clipboard: false,
                    content: './ui/assets',
                    port: 3080,
                    hotClient: false,
                    devMiddleware: {
                        publicPath: '/.assets/',
                        stats: WEBPACK_STATS_OPTIONS,
                    },
                    add: (app, middleware) => {
                        // Since we're manipulating the order of middleware added, we need to handle adding these
                        // two internal middleware functions.
                        //
                        // The `as any` cast is necessary because the `middleware.webpack` typings are incorrect
                        // (the related issue https://github.com/webpack-contrib/webpack-serve/issues/238 perhaps
                        // explains why: the webpack-serve docs incorrectly state that resolving
                        // `middleware.webpack()` is not necessary).
                        ;(middleware.webpack() as any).then(() => {
                            middleware.content()

                            // Proxy *must* be the last middleware added.
                            app.use(
                                convert(
                                    // Proxy all requests (that are not for webpack-built assets) to the Sourcegraph
                                    // frontend server, and we make the Sourcegraph appURL equal to the URL of
                                    // webpack-serve. This is how webpack-serve needs to work (because it does a bit
                                    // more magic in injecting scripts that use WebSockets into proxied requests).
                                    httpProxyMiddleware({
                                        target: 'http://localhost:3081',
                                        ws: true,

                                        // Avoid crashing on "read ECONNRESET".
                                        onError: err => console.error(err),
                                        onProxyReqWs: (_proxyReq, _req, socket) =>
                                            socket.on('error', err => console.error('WebSocket proxy error:', err)),
                                    })
                                )
                            )
                        })
                    },
                    compiler: createWebpackCompiler(webpackConfig),
                },
            },
        }
    )
}

const GRAPHQL_SCHEMA_PATH = __dirname + '/cmd/frontend/graphqlbackend/schema.graphql'

export async function watchGraphQLTypes(): Promise<void> {
    await new Promise<never>((resolve, reject) => {
        gulp.watch(GRAPHQL_SCHEMA_PATH, graphQLTypes).on('error', reject)
    })
}

/** Generates the TypeScript types for the GraphQL schema */
export async function graphQLTypes(): Promise<void> {
    const schemaStr = await readFile(GRAPHQL_SCHEMA_PATH, 'utf8')
    const schema = buildSchema(schemaStr)
    const result = (await graphql(schema, introspectionQuery)) as { data: IntrospectionQuery }

    const formatOptions = (await resolveConfig(__dirname, { config: __dirname + '/prettier.config.js' }))!
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
    await writeFile(__dirname + '/src/backend/graphqlschema.ts', typings)
}

/**
 * Generates the TypeScript types for the JSON schemas and copies the schemas to src/ so they can be imported
 */
export async function schema(): Promise<void> {
    await Promise.all([mkdirp(__dirname + '/src/schema'), mkdirp(__dirname + '/dist/schema')])
    await Promise.all(
        ['json-schema', 'settings', 'site', 'extension'].map(async file => {
            let schema = await readFile(__dirname + `/schema/${file}.schema.json`, 'utf8')
            // HACK: Rewrite absolute $refs to be relative. They need to be absolute for Monaco to resolve them
            // when the schema is in a oneOf (to be merged with extension schemas).
            schema = schema.replace(
                /https:\/\/sourcegraph\.com\/v1\/settings\.schema\.json#\/definitions\//g,
                '#/definitions/'
            )
            const types = await compileJSONSchema(JSON.parse(schema), 'settings.schema', {
                cwd: __dirname + '/schema',
            })
            await Promise.all([
                writeFile(__dirname + `/src/schema/${file}.schema.d.ts`, types),
                // Copy schema to src/ so it can be imported in TypeScript
                writeFile(__dirname + `/src/schema/${file}.schema.json`, schema),
                // Copy schema to dist/ so it's part of the dist package
                // This would not be needed with tsconfig `resolevJsonModule: true`,
                // but we cannot enable that because of https://github.com/Microsoft/TypeScript/issues/25755
                // and TS3.1 has blocking compiler bugs
                writeFile(__dirname + `/dist/schema/${file}.schema.json`, schema),
            ])
        })
    )
}

export async function watchSchema(): Promise<void> {
    await new Promise<never>((resolve, reject) => {
        gulp.watch(__dirname + '/schema/*.schema.json', schema).on('error', reject)
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

/**
 * Builds and typechecks the TypeScript code, outputting compiled JavaScript, declaration files and sourcemaps to dist/
 */
export function typescript(): ChildProcess {
    return spawn(__dirname + '/node_modules/.bin/tsc', ['-p', __dirname + '/tsconfig.dist.json', '--pretty'], {
        stdio: 'inherit',
        shell: true,
    })
}

export function watchTypescript(): ChildProcess {
    return spawn(
        __dirname + '/node_modules/.bin/tsc',
        ['-p', __dirname + '/tsconfig.dist.json', '--watch', '--preserveWatchOutput', '--pretty'],
        {
            stdio: 'inherit',
            shell: true,
        }
    )
}

const SASS_FILES = './src/**/*.scss'

/**
 * Copies the .scss files from src/ to dist/.
 * These are not precompiled so that they can be imported individually and variables be set.
 */
export function sass(): NodeJS.ReadWriteStream {
    return gulp.src(SASS_FILES).pipe(gulp.dest('./dist'))
}

export const watchSass = gulp.series(sass, async function watchSass(): Promise<void> {
    await new Promise<never>((_, reject) => {
        gulp.watch(SASS_FILES, sass).on('error', reject)
    })
})

/**
 * Builds only the dist/ folder.
 */
export const dist = gulp.parallel(sass, gulp.series(gulp.parallel(schema, graphQLTypes), typescript))
export const watchDist = gulp.series(
    // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
    gulp.parallel(schema, graphQLTypes),
    gulp.parallel(watchSass, watchSchema, watchGraphQLTypes, watchTypescript)
)

/**
 * Builds everything.
 */
export const build = gulp.parallel(
    sass,
    gulp.series(gulp.parallel(schema, graphQLTypes), gulp.parallel(webpack, typescript))
)

/**
 * Watches everything and rebuilds on file changes.
 */
export const watch = gulp.series(
    // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
    gulp.parallel(schema, graphQLTypes),
    gulp.parallel(watchSass, watchSchema, watchGraphQLTypes, watchTypescript, webpackServe)
)

/**
 * Publishes a new version of @sourcegraph/webapp to npm.
 * Gets the last release from the npm registry, increases the patch version, writes it to package.json and publishes the package.
 * It is not a goal to parse commit messages or follow semantic versioning - every commit gets released as a 0.0.x release.
 * No git tags or GitHub releases are created.
 */
export async function release(): Promise<void> {
    const packageJson = require('./package.json')
    try {
        const currentVersion = await latestVersion(packageJson.name)
        log(`Current version is ${currentVersion}`)
        packageJson.version = semver.inc(currentVersion, 'patch')
    } catch (err) {
        if (/doesn't exist/.test(err.message)) {
            log('Package is not released yet')
            packageJson.version = '0.0.0'
        } else {
            throw err
        }
    }
    log(`New version is ${packageJson.version}`)
    if (!process.env.CI) {
        log('Not running in CI, aborting')
        return
    }
    await writeFile(__dirname + '/package.json', JSON.stringify(packageJson, null, 2))
    await execa('npm', ['publish'], { stdio: 'inherit' })
    await execa('buildkite-agent', ['meta-data', 'set', 'oss-webapp-version', packageJson.version])
}
