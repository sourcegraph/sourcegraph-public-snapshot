import { generateNamespace } from '@gql2ts/from-schema'
import { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } from '@gql2ts/language-typescript'
import { buildSchema, graphql, introspectionQuery, IntrospectionQuery } from 'graphql'
import gulp from 'gulp'
import $RefParser from 'json-schema-ref-parser'
import { compile as compileJSONSchema } from 'json-schema-to-typescript'
import mkdirp from 'mkdirp-promise'
import { readFile, writeFile } from 'mz/fs'
import path from 'path'
import { format, resolveConfig } from 'prettier'
import { draftV7resolver } from './draftV7Resolver'

const GRAPHQL_SCHEMA_PATH = path.join(__dirname, '../cmd/frontend/graphqlbackend/schema.graphql')

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
    await writeFile(__dirname + '/src/graphql/schema.ts', typings)
}

/**
 * Generates the TypeScript types for the JSON schemas.
 */
export async function schema(): Promise<void> {
    const outputDir = path.join(__dirname, '..', 'web', 'src', 'schema')
    await mkdirp(outputDir)
    const schemaDir = path.join(__dirname, '..', 'schema')
    await Promise.all(
        ['json-schema-draft-07', 'settings', 'critical', 'site'].map(async file => {
            let schema = await readFile(path.join(schemaDir, `${file}.schema.json`), 'utf8')
            // HACK: Rewrite absolute $refs to be relative. They need to be absolute for Monaco to resolve them
            // when the schema is in a oneOf (to be merged with extension schemas).
            schema = schema.replace(
                /https:\/\/sourcegraph\.com\/v1\/settings\.schema\.json#\/definitions\//g,
                '#/definitions/'
            )

            const types = await compileJSONSchema(JSON.parse(schema), 'settings.schema', {
                cwd: schemaDir,
                $refOptions: {
                    resolve: {
                        draftV7resolver,
                        // there should be no reason to make network calls during this process,
                        // and if there are we've broken env for offline devs/increased dev startup time
                        http: false,
                    } as $RefParser.Options['resolve'],
                },
            })
            await writeFile(path.join(outputDir, `${file}.schema.d.ts`), types)
        })
    )
}

export async function watchSchema(): Promise<void> {
    await new Promise<never>((_resolve, reject) => {
        gulp.watch(__dirname + '/../schema/*.schema.json', schema).on('error', reject)
    })
}
