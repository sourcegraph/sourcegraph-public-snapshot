/* eslint-disable no-template-curly-in-string */
import assert from 'assert'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'
import { commonWebGraphQlResults } from './graphQlResults'
import { RegistryExtensionFieldsForList } from '../graphql-operations'
import { siteGQLID, siteID } from './jscontext'
import { ExtensionsResult } from '../../../shared/src/graphql-operations'
// import { isExtensionEnabled } from '../../../shared/src/extensions/extension'

const typescriptRawManifest =
    '{\n  "$schema": "https://raw.githubusercontent.com/sourcegraph/sourcegraph/master/shared/src/schema/extension.schema.json",\n  "activationEvents": [\n    "onLanguage:typescript",\n    "onLanguage:javascript"\n  ],\n  "browserslist": [\n    "last 1 Chrome versions",\n    "last 1 Firefox versions",\n    "last 1 Edge versions",\n    "last 1 Safari versions"\n  ],\n  "categories": [\n    "Programming languages"\n  ],\n  "contributes": {\n    "actions": [\n      {\n        "actionItem": {\n          "label": "${config.typescript.showExternalReferences \\u0026\\u0026 \\"Hide references from other repositories\\" || \\"Show references from other repositories\\"}"\n        },\n        "command": "updateConfiguration",\n        "commandArguments": [\n          [\n            "typescript.showExternalReferences"\n          ],\n          "${!config.typescript.showExternalReferences}",\n          null,\n          "json"\n        ],\n        "id": "externalReferences.toggle",\n        "title": "${config.typescript.showExternalReferences \\u0026\\u0026 \\"Hide references from other repositories\\" || \\"Show references from other repositories\\"}"\n      }\n    ],\n    "configuration": {\n      "additionalProperties": false,\n      "properties": {\n        "basicCodeIntel.includeArchives": {\n          "description": "Whether to include archived repositories in search results.",\n          "type": "boolean"\n        },\n        "basicCodeIntel.includeForks": {\n          "description": "Whether to include forked repositories in search results.",\n          "type": "boolean"\n        },\n        "basicCodeIntel.indexOnly": {\n          "description": "Whether to use only indexed requests to the search API.",\n          "type": "boolean"\n        },\n        "basicCodeIntel.unindexedSearchTimeout": {\n          "description": "The timeout (in milliseconds) for un-indexed search requests.",\n          "type": "number"\n        },\n        "codeIntel.lsif": {\n          "description": "Whether to use pre-computed LSIF data for code intelligence (such as hovers, definitions, and references). See https://docs.sourcegraph.com/user/code_intelligence/lsif.",\n          "type": "boolean"\n        },\n        "typescript.accessToken": {\n          "description": "The access token for the language server to use to fetch files from the Sourcegraph API. The extension will create this token and save it in your settings automatically.",\n          "type": "string"\n        },\n        "typescript.diagnostics.enable": {\n          "description": "Whether to show compile errors on lines (Default: false)",\n          "type": "boolean"\n        },\n        "typescript.langserver.log": {\n          "description": "The log level to pass to the TypeScript language server. Logs will be forwarded to the browser console with the prefix [langserver].",\n          "enum": [\n            false,\n            "log",\n            "info",\n            "warn",\n            "error"\n          ]\n        },\n        "typescript.maxExternalReferenceRepos": {\n          "description": "The maximum number of dependent packages to look in when searching for external references for a symbol (defaults to 20).",\n          "type": "number"\n        },\n        "typescript.npmrc": {\n          "description": "Settings to be written into an npmrc in key/value format. Can be used to specify custom registries and tokens.",\n          "type": "object"\n        },\n        "typescript.progress": {\n          "description": "Whether to report progress while fetching sources, installing dependencies etc. (Default: true)",\n          "type": "boolean"\n        },\n        "typescript.restartAfterDependencyInstallation": {\n          "description": "Whether to restart the language server after dependencies were installed (default true)",\n          "type": "boolean"\n        },\n        "typescript.serverUrl": {\n          "description": "The address of the WebSocket language server to connect to (e.g. ws://host:8080).",\n          "format": "url",\n          "type": "string"\n        },\n        "typescript.showExternalReferences": {\n          "description": "Whether or not a second references provider for external references will be registered (defaults to false).",\n          "type": "boolean"\n        },\n        "typescript.sourcegraphUrl": {\n          "description": "The address of the Sourcegraph instance from the perspective of the TypeScript language server.",\n          "format": "url",\n          "type": "string"\n        },\n        "typescript.tsserver.log": {\n          "description": "The log level to pass to tsserver. Logs will be forwarded to the browser console with the prefix [tsserver].",\n          "enum": [\n            false,\n            "terse",\n            "normal",\n            "requestTime",\n            "verbose"\n          ]\n        }\n      },\n      "title": "Settings",\n      "type": "object"\n    },\n    "menus": {\n      "panel/toolbar": [\n        {\n          "action": "externalReferences.toggle",\n          "when": "panel.activeView.id == \'references\'"\n        }\n      ]\n    }\n  },\n  "description": "TypeScript/JavaScript code intelligence",\n  "engines": {\n    "node": "\\u003e=11.1.0"\n  },\n  "extensionID": "sourcegraph/typescript",\n  "files": [\n    "dist"\n  ],\n  \n  "license": "MIT",\n  "main": "dist/extension.js",\n  "name": "typescript",\n  "private": true,\n  "publisher": "sourcegraph",\n  "readme": "# Code intelligence for TypeScript/JavaScript\\n\\nThis extension provides TypeScript/JavaScript code intelligence on Sourcegraph.\\n\\n[**🗃️ Source code**](https://github.com/sourcegraph/code-intel-extensions/tree/master/extensions/typescript)\\n\\n![TypeScript code intelligence](https://user-images.githubusercontent.com/133014/63376874-a92c7900-c343-11e9-98bb-631016f1eff7.gif)\\n",\n  "repository": {\n    "directory": "extensions/typescript",\n    "type": "git",\n    "url": "github:sourcegraph/code-intel-extensions"\n  },\n  "scripts": {\n    "build": "tsc -b . \\u0026\\u0026 parcel build --out-file extension.js src/extension.ts \\u0026\\u0026 dot-json dist/extension.map sourceRoot https://sourcegraph.com/github.com/sourcegraph/sourcegraph-typescript@$(git rev-parse HEAD)/-/raw/extension/src/",\n    "config-types": "dot-json package.json contributes.configuration | json2ts --unreachableDefinitions --style.singleQuote --no-style.semi -o src/settings.ts",\n    "publish": "yarn -s \\u0026\\u0026 src -config \\"${SRC_CONFIG}\\" ext publish",\n    "serve": "yarn run symlink-package \\u0026\\u0026 parcel serve --no-hmr --out-file dist/extension.js src/extension.ts",\n    "sourcegraph:prepublish": "yarn run build",\n    "symlink-package": "mkdirp dist \\u0026\\u0026 lnfs ./package.json ./dist/package.json"\n  },\n  "sideEffects": false,\n  "tags": [\n    "typescript",\n    "javascript",\n    "cross-repository",\n    "language-server",\n    "react",\n    "jsx"\n  ],\n  "url": "https://sourcegraph.com/-/static/extension/7745-sourcegraph-typescript.js?c4ort27pnr00--sourcegraph-typescript",\n  "version": "0.0.0-DEVELOPMENT"\n}'

const registryExtensionNodes: RegistryExtensionFieldsForList[] = [
    {
        id: 'test-extension-1',
        publisher: null,
        extensionID: 'sourcegraph/typescript',
        extensionIDWithoutRegistry: 'sourcegraph/typescript',
        name: 'typescript',
        manifest: {
            raw: typescriptRawManifest,
            description: 'TypeScript/JavaScript code intelligence',
        },
        createdAt: '2019-01-26T03:39:17Z',
        updatedAt: '2019-01-26T03:39:17Z',
        url: '/extensions/sourcegraph/typescript',
        remoteURL: 'https://sourcegraph.com/extensions/sourcegraph/typescript',
        registryName: 'sourcegraph.com',
        isLocal: false,
        isWorkInProgress: false,
        viewerCanAdminister: false,
    },
]

const extensionNodes: ExtensionsResult['extensionRegistry']['extensions']['nodes'] = [
    {
        extensionID: 'sourcegraph/typescript',
        id: 'test-extension-1',
        manifest: {
            raw: typescriptRawManifest,
        },
        url: '/extensions/sourcegraph/typescript',
        viewerCanAdminister: false,
    },
]

describe('Extension Registry', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    saveScreenshotsUponFailures(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('toggling', () => {
        function overrideGraphQLForExtension({ enabled }: { enabled: boolean }): void {
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ViewerSettings: () => ({
                    viewerSettings: {
                        subjects: [
                            {
                                __typename: 'DefaultSettings',
                                settingsURL: null,
                                viewerCanAdminister: false,
                                latestSettings: {
                                    id: 0,
                                    contents: JSON.stringify({}),
                                },
                            },
                            {
                                __typename: 'Site',
                                id: siteGQLID,
                                siteID,
                                latestSettings: {
                                    id: 470,
                                    contents: JSON.stringify({}),
                                },
                                settingsURL: '/site-admin/global-settings',
                                viewerCanAdminister: true,
                            },
                            {
                                __typename: 'User',
                                id: 'TestGQLUserID',
                                username: 'testusername',
                                settingsURL: '/user/testusername/settings',
                                displayName: 'test',
                                viewerCanAdminister: true,
                                latestSettings: {
                                    id: 310,
                                    contents: JSON.stringify({ extensions: { 'sourcegraph/typescript': enabled } }),
                                },
                            },
                        ],
                        final: JSON.stringify({}),
                    },
                }),
                RegistryExtensions: () => ({
                    extensionRegistry: {
                        extensions: {
                            error: null,
                            nodes: registryExtensionNodes,
                        },
                    },
                }),
                Extensions: () => ({
                    extensionRegistry: {
                        extensions: {
                            nodes: extensionNodes,
                        },
                    },
                }),
                EditSettings: () => ({
                    configurationMutation: {
                        editConfiguration: {
                            empty: null,
                        },
                    },
                }),
            })
        }

        it('a disabled extension enables it', async () => {
            const enabled = false
            overrideGraphQLForExtension({ enabled })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            // toggle typescript extension on
            const { edit } = await testContext.waitForGraphQLRequest(async () => {
                const toggle = await driver.page.waitForSelector("[id='test-extension-toggle-sourcegraph/typescript']")
                await toggle.click()
            }, 'EditSettings')

            assert.deepStrictEqual(edit, {
                keyPath: [{ property: 'extensions' }, { property: 'sourcegraph/typescript' }],
                value: !enabled,
            })
        })

        it('an enabled extension disables it ', async () => {
            const enabled = true
            overrideGraphQLForExtension({ enabled })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            // toggle typescript extension off
            const { edit } = await testContext.waitForGraphQLRequest(async () => {
                const toggle = await driver.page.waitForSelector("[id='test-extension-toggle-sourcegraph/typescript']")
                await toggle.click()
            }, 'EditSettings')

            assert.deepStrictEqual(edit, {
                keyPath: [{ property: 'extensions' }, { property: 'sourcegraph/typescript' }],
                value: !enabled,
            })
        })
    })
})
