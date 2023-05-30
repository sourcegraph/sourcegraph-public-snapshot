/**
 * This module defines shared VSCode mocks for use in every Jest test.
 * Tests requiring no custom mocks will automatically apply the mocks defined in this file.
 * This is made possible via the `setupFilesAfterEnv` property in the Jest configuration.
 */

class Position {
    public line: number
    public character: number

    constructor(line: number, character: number) {
        this.line = line
        this.character = character
    }
}

class Range {
    public startLine?: number
    public startCharacter?: number
    public endLine?: number
    public endCharacter?: number
    public start: Position
    public end: Position

    constructor(
        startLine: number | Position,
        startCharacter: number | Position,
        endLine: number,
        endCharacter: number
    ) {
        if (typeof startLine !== 'number' && typeof startCharacter !== 'number') {
            this.start = startLine
            this.end = startCharacter
        } else if (typeof startLine === 'number' && typeof startCharacter === 'number') {
            this.startLine = startLine
            this.startCharacter = startCharacter
            this.endLine = endLine
            this.endCharacter = endCharacter
            this.start = new Position(startLine, startCharacter)
            this.end = new Position(endLine, endCharacter)
        } else {
            throw new TypeError('this version of the constructor is not implemented')
        }
    }
}

class InlineCompletionItem {
    public content: string
    constructor(content: string) {
        this.content = content
    }
}

const vsCodeMocks = {
    Range,
    Position,
    InlineCompletionItem,
    window: {
        showInformationMessage: () => undefined,
        showWarningMessage: () => undefined,
        showQuickPick: () => undefined,
        showInputBox: () => undefined,
        createOutputChannel() {
            return null
        },
        showErrorMessage(message: string) {
            throw new Error(message)
        },
        activeTextEditor: { document: { uri: { scheme: 'not-cody' } } },
    },
    workspace: {
        getConfiguration() {
            return undefined
        },
    },
    ConfigurationTarget: {
        Global: undefined,
    },
} as const

/**
 * Mock name is required to keep Jest happy and avoid the error:
 * "The module factory of jest.mock() is not allowed to reference any out-of-scope variables"
 *
 * This function can be used to customize the default VSCode mocks in any test file.
 */
export function mockVSCodeExports(): typeof vsCodeMocks {
    return vsCodeMocks
}

/**
 * Apply the default VSCode mocks to the global scope.
 */
jest.mock('vscode', () => mockVSCodeExports())
