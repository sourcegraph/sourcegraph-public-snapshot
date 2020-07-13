/* eslint-disable unicorn/consistent-function-scoping */
/* eslint-disable no-sync */
import * as prettier from 'prettier'
import { Extractor } from 'ts-graphql-plugin/lib/analyzer/extractor'
import { ScriptHost } from 'ts-graphql-plugin/lib/analyzer/analyzer-factory'
import { createScriptSourceHelper } from 'ts-graphql-plugin/lib/ts-ast-util/script-source-helper'
import * as path from 'path'
import * as fs from 'fs'
import { buildSchema, GraphQLSchema } from 'graphql'
import * as ts from 'typescript'
import { TsGraphQLPluginConfigOptions } from 'ts-graphql-plugin/lib/types'
import { extractTypes, createTsTypeDeclaration } from './gql2ts-transformer'
import { memoize } from 'lodash'

// NOTE borderline reasonable to define an interface here instead of duplicating
const WEB_TSPROJECT_PATH = path.join(__dirname, '../../web')
const WEB_OUTPUT_DIR = path.join(__dirname, '../../web/src')
const WEB_INTERFACE_NAME = 'WebGQLOperations'

const SHARED_TSPROJECT_PATH = path.join(__dirname, '../')
const SHARED_OUTPUT_DIR = path.join(__dirname, '../src/')
const SHARED_INTERFACE_NAME = 'SharedGQLOperations'

const readSchema = (schemaPath: string): GraphQLSchema => {
    const isExists = fs.existsSync(schemaPath)
    if (!isExists) {
        throw new Error('schema file was not found here:' + schemaPath)
    }

    const sdl = fs.readFileSync(schemaPath, 'utf-8')
    const schema = buildSchema(sdl)
    return schema
}

export async function formatSourceFile(fileContent: string, fileLocation: string): Promise<string> {
    const config = await prettier.resolveConfig(fileLocation)
    if (!config) {
        throw new Error(`Prettier config not found for file ${fileLocation}`)
    }
    return prettier.format(fileContent, { ...config, parser: 'typescript' })
}

const extractGQL = async (tsProjectPath: string, outputDirectory: string, interfaceName: string): Promise<void> => {
    const { pluginConfig, tsconfig, prjRootPath } = readTsconfig(tsProjectPath)
    if (typeof pluginConfig.schema !== 'string') {
        throw new TypeError('for now schema field needs to be a string path')
    }

    const schema = readSchema(path.join(prjRootPath, pluginConfig.schema))
    const currentDirectory = process.cwd()
    const scriptHost = new ScriptHost(currentDirectory, tsconfig.options)
    for (const fileName of tsconfig.fileNames) {
        scriptHost.readFile(fileName)
    }

    const langService = ts.createLanguageService(scriptHost)
    const scriptSourceHelper = createScriptSourceHelper({
        languageService: langService,
        languageServiceHost: scriptHost,
    })

    const extractor = new Extractor({
        removeDuplicatedFragments: true,
        scriptSourceHelper,
        debug: message => console.log('##gql: ' + message),
    })

    const extractedResults = extractor.extract(scriptHost.getScriptFileNames(), pluginConfig.tag)

    let typeDeclarations: ts.Statement[] = []
    const members: ts.PropertySignature[] = []

    const operationInputName = (name: string): string => name + 'Variables'
    const operationOutputName = (name: string): string => name + 'Result'

    const fragmentNames = new Set<string>()

    for (const result of extractedResults) {
        if (!result.documentNode) {
            continue
        }

        const { type } = extractor.getDominantDefiniton(result)
        if (type === 'complex') {
            throw new Error('not complex types')
        }

        // TODO this is a hack to generate unique names for fragments
        // note within a single document it should be memoized
        const uniqueFragmentName = memoize((name: string): string => {
            if (!fragmentNames.has(name)) {
                fragmentNames.add(name)
                return name
            }
            let index = 1
            while (true) {
                const potentialName = name + index.toString()
                if (fragmentNames.has(potentialName)) {
                    index += 1
                    continue
                }

                fragmentNames.add(potentialName)
                return potentialName
            }
        })

        const types = extractTypes(result.documentNode, result.fileName, schema, uniqueFragmentName)
        for (const type of types) {
            if (type.tag === 'fragment') {
                typeDeclarations.push(createTsTypeDeclaration(type.name, type.output))
            }
            if (type.tag === 'operation') {
                const inputTypeName = operationInputName(type.name)
                typeDeclarations.push(createTsTypeDeclaration(inputTypeName, type.input))
                const outputTypeName = operationOutputName(type.name)
                typeDeclarations.push(createTsTypeDeclaration(outputTypeName, type.output))

                // generates something like updateBla: (input: BlaInput) => BlaOutput
                members.push(
                    ts.createPropertySignature(
                        undefined,
                        type.name + `/* ${path.relative(prjRootPath, result.fileName)} */`,
                        undefined,
                        ts.createFunctionTypeNode(
                            undefined,
                            [
                                ts.createParameter(
                                    undefined,
                                    undefined,
                                    undefined,
                                    'variables',
                                    undefined,
                                    ts.createTypeReferenceNode(inputTypeName, undefined)
                                ),
                            ],
                            ts.createTypeReferenceNode(outputTypeName, undefined)
                        ),
                        undefined
                    )
                )
            }
        }
    }

    typeDeclarations = [
        ts.createInterfaceDeclaration(
            undefined,
            ts.createModifiersFromModifierFlags(ts.ModifierFlags.Export),
            interfaceName,
            undefined,
            undefined,
            members
        ),
        ...typeDeclarations,
    ]

    // TODO as an option
    const outputFileName = 'gql-operations.ts'

    const sourceFile = ts.createSourceFile(outputFileName, '', ts.ScriptTarget.Latest, false, ts.ScriptKind.TS)
    const resultFile = ts.updateSourceFileNode(sourceFile, typeDeclarations)

    const comments = [
        'eslint-disable @typescript-eslint/consistent-type-definitions',
        'This is an autogenerated file. Do not edit this file directly!',
    ]
    for (const comment of comments) {
        ts.addSyntheticLeadingComment(
            resultFile.statements[0],
            ts.SyntaxKind.MultiLineCommentTrivia,
            ` ${comment} `,
            true
        )
    }

    const outputFilePath = path.join(outputDirectory, outputFileName)
    const printer = ts.createPrinter({ newLine: ts.NewLineKind.LineFeed, removeComments: false })
    const prettySource = await formatSourceFile(printer.printFile(resultFile), outputFilePath)

    ts.sys.writeFile(outputFilePath, prettySource)
}

interface Config {
    tsconfig: ts.ParsedCommandLine
    pluginConfig: TsGraphQLPluginConfigOptions
    prjRootPath: string
}

// copy pasted
function readTsconfig(project: string): Config {
    const currentDirectory = ts.sys.getCurrentDirectory()
    const ppath = path.isAbsolute(project) ? path.resolve(currentDirectory, project) : project
    let configPath: string | undefined
    if (ts.sys.fileExists(ppath)) {
        configPath = ppath
    } else if (ts.sys.directoryExists(ppath) && ts.sys.fileExists(path.join(ppath, 'tsconfig.json'))) {
        configPath = path.join(ppath, 'tsconfig.json')
    }
    if (!configPath) {
        throw new Error(`tsconfig file not found: ${project}`)
    }
    const tsconfig = ts.getParsedCommandLineOfConfigFile(configPath, {}, ts.sys as any)
    if (!tsconfig) {
        throw new Error(`Failed to parse: ${configPath}`)
    }
    const prjRootPath = path.dirname(configPath)
    const plugins = tsconfig.options.plugins
    if (!plugins || !Array.isArray(plugins)) {
        throw new Error(
            `tsconfig.json should have ts-graphql-plugin setting. Add the following:
  "compilerOptions": {
    "plugins": [
      {
        "name": "ts-graphql-plugin",
        "schema": "shema.graphql",   /* Path to your GraphQL schema */
        "tag": "gql"                 /* Template tag function name */
      }
    ]
  }`
        )
    }
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    // eslint-disable-next-line id-length
    const found = (plugins as any[]).find((p: any) => p.name === 'ts-graphql-plugin')
    if (!found) {
        throw new Error(
            `tsconfig.json should have ts-graphql-plugin setting. Add the following:
  "compilerOptions": {
    "plugins": [
      {
        "name": "ts-graphql-plugin",
        "schema": "shema.graphql",   /* Path to your GraphQL schema */
        "tag": "gql"                 /* Template tag function name */
      }
    ]
  }`
        )
    }
    const pluginConfig = found as TsGraphQLPluginConfigOptions
    return {
        tsconfig,
        pluginConfig,
        prjRootPath,
    }
}

Promise.all([
    // web
    extractGQL(WEB_TSPROJECT_PATH, WEB_OUTPUT_DIR, WEB_INTERFACE_NAME),
    // shared
    extractGQL(SHARED_TSPROJECT_PATH, SHARED_OUTPUT_DIR, SHARED_INTERFACE_NAME),
]).catch(error => {
    console.error('Error happened during extraction', error)
    process.exit(1)
})
