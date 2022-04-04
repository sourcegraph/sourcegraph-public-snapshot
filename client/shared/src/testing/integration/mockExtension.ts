import { PollyServer } from '@pollyjs/core'
import type * as sourcegraph from 'sourcegraph'

import { ExtensionManifest } from '../../extensions/extensionManifest'
import { ExtensionsResult, SharedGraphQlOperations } from '../../graphql-operations'
import { Settings } from '../../settings/settings'

interface ExtensionMockingInit {
    /**
     * The polly server object, used to intercept extension bundle requests.
     */
    pollyServer: PollyServer
    /**
     * The base Sourcegraph URL for the test instance, used to construst bundle URL.
     */
    sourcegraphBaseUrl: string
}

interface ExtensionMockingUtils {
    /**
     * Adds, enables, and sets up request interception for an extension with the given ID.
     * Bundle is a function that imports "sourcegraph" with `require` (i.e. `const sourcegraph = require('sourcegraph');`)
     * and exports an `activate` function, just like any other Sourcegraph extension.
     */
    mockExtension: ({ id, bundle }: { id: string; bundle: () => void }) => void
    /**
     * Use this as the `Extension` override for `TestContext#overrideGraphQL`.
     */
    Extensions: SharedGraphQlOperations['Extensions']
    /**
     * Merge/replace your mock settings `extensions` property with this object.
     */
    extensionSettings: Settings['extensions']
}

interface ExtensionsResultMock {
    extensionRegistry: ExtensionsResult['extensionRegistry'] & {
        __typename: 'ExtensionRegistry'
    }
}

/**
 * Set up Sourcegraph extension mocking for an integration test.
 */
export function setupExtensionMocking({
    pollyServer,
    sourcegraphBaseUrl,
}: ExtensionMockingInit): ExtensionMockingUtils {
    let internalID = 0

    const extensionSettings: Settings['extensions'] = {}
    const extensionsResult: ExtensionsResultMock = {
        extensionRegistry: {
            __typename: 'ExtensionRegistry',
            extensions: {
                nodes: [],
            },
        },
    }

    return {
        mockExtension: ({ id, bundle }) => {
            internalID++

            /** The URL at which the manifest says the extension bundle is served. We should intercept requests to this URL. */
            const bundleURL = new URL(
                `/-/static/extension/00${internalID}-${id.replace(/\//g, '-')}.js?hash--${id.replace(/\//g, '-')}`,
                sourcegraphBaseUrl
            ).href

            const extensionManifest: ExtensionManifest = {
                url: bundleURL,
                activationEvents: ['*'],
            }

            // Mutate mock data objects
            extensionSettings[id] = true
            extensionsResult.extensionRegistry.extensions.nodes.push({
                id,
                extensionID: id,
                manifest: {
                    jsonFields: extensionManifest,
                },
            })

            pollyServer.get(bundleURL).intercept((request, response) => {
                // Create an immediately-invoked function expression for the extensionBundle function
                const extensionBundleString = `(${bundle.toString()})()`
                response.type('application/javascript; charset=utf-8').send(extensionBundleString)
            })
        },
        Extensions: () => extensionsResult,
        extensionSettings,
    }
}

// Commonly mocked extensions.

/**
 * A simple hover provider extension.
 * Shows the token that the user is hovering over in the hover overlay.
 * Used to verify that the correct document and position info reaches the extension.
 */
export function simpleHoverProvider(): void {
    // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
    const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

    function activate(context: sourcegraph.ExtensionContext): void {
        context.subscriptions.add(
            sourcegraph.languages.registerHoverProvider(['*'], {
                provideHover: (document, position) => {
                    const range = document.getWordRangeAtPosition(position)
                    const token = document.getText(range)
                    if (!token) {
                        return null
                    }
                    return {
                        contents: {
                            value: `User is hovering over ${token}`,
                            kind: sourcegraph.MarkupKind.Markdown,
                        },
                        range,
                    }
                },
            })
        )
    }

    exports.activate = activate
}
