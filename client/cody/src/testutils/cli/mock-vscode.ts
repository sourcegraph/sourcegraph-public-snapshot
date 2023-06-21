import mock from 'mock-require'

mock('vscode', {})

// We need to mock `vscode` first, so we use require here.
// eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
const { vsCodeMocks } = require('../mocks')

const vscodeMock = {
    ...vsCodeMocks,
    InlineCompletionTriggerKind: {
        Invoke: 0,
        Automatic: 1,
    },
    workspace: {
        asRelativePath(path: string) {
            return path
        },
        getConfiguration() {
            return {
                get(key: string) {
                    switch (key) {
                        case 'cody.debug.filter':
                            return '.*'
                        case 'cody.experimental.suggestions':
                            return true
                        case 'cody.serverEndpoint':
                            return 'https://sourcegraph.sourcegraph.com'
                        default:
                            return undefined
                    }
                },
            }
        },
        onDidChangeTextDocument() {
            return null
        },
    },
}

// This will replace any require('vscode') with our mock
mock('vscode', vscodeMock)
