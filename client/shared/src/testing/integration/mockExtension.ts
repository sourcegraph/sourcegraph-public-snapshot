import type { ExtensionContext } from '../../codeintel/legacy-extensions/api'

interface ExtensionMockingUtils {
    /**
     * Adds, enables, and sets up request interception for an extension with the given ID.
     * Bundle is a function that imports "sourcegraph" with `require` (i.e. `const sourcegraph = require('sourcegraph');`)
     * and exports an `activate` function, just like any other Sourcegraph extension.
     */
    mockExtension: ({ id, bundle }: { id: string; bundle: () => void }) => void
    /**
     * Merge/replace your mock settings `extensions` property with this object.
     */
    extensionSettings: {}
}

/**
 * Set up Sourcegraph extension mocking for an integration test.
 */
export function setupExtensionMocking(): ExtensionMockingUtils {
    throw new Error('not yet reimplemented after the extension API deprecation')
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

    function activate(context: ExtensionContext): void {
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
