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

    public isAfter(other: Position): boolean {
        return other.line < this.line || (other.line === this.line && other.character < this.character)
    }
    public isAfterOrEqual(other: Position): boolean {
        return this.isAfter(other) || this.isEqual(other)
    }
    public isBefore(other: Position): boolean {
        return !this.isAfterOrEqual(other)
    }
    public isBeforeOrEqual(other: Position): boolean {
        return !this.isAfter(other)
    }
    public isEqual(other: Position): boolean {
        return this.line === other.line && this.character === other.character
    }
    public translate(lineDelta?: number, characterDelta?: number): Position {
        return new Position(this.line + (lineDelta || 0), this.character + (characterDelta || 0))
    }
}

class Range {
    public start: Position
    public end: Position

    constructor(
        startLine: number | Position,
        startCharacter: number | Position,
        endLine?: number,
        endCharacter?: number
    ) {
        if (typeof startLine !== 'number' && typeof startCharacter !== 'number') {
            this.start = startLine
            this.end = startCharacter
        } else if (
            typeof startLine === 'number' &&
            typeof startCharacter === 'number' &&
            typeof endLine === 'number' &&
            typeof endCharacter === 'number'
        ) {
            this.start = new Position(startLine, startCharacter)
            this.end = new Position(endLine, endCharacter)
        } else {
            throw new TypeError('this version of the constructor is not implemented')
        }
    }

    public with(start: Position, end: Position): Range {
        return start.isEqual(this.start) && end.isEqual(this.end) ? this : new Range(start, end)
    }
    public get startLine(): number {
        return this.start.line
    }
    public get startCharacter(): number {
        return this.start.character
    }
    public get endLine(): number {
        return this.end.line
    }
    public get endCharacter(): number {
        return this.end.character
    }
}

class Uri {
    public fsPath: string
    public path: string
    constructor(path: string) {
        this.fsPath = path
        this.path = path
    }
}

class InlineCompletionItem {
    public insertText: string
    constructor(content: string) {
        this.insertText = content
    }
}

// TODO(abeatrix): Implement delete and insert mocks
class WorkspaceEdit {
    public delete(uri: Uri, range: Range): Range {
        return range
    }
    public insert(uri: Uri, position: Position, content: string): string {
        return content
    }
}

const vsCodeMocks = {
    Range,
    Position,
    InlineCompletionItem,
    WorkspaceEdit,
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
        activeTextEditor: { document: { uri: { scheme: 'not-cody' } }, options: { tabSize: 4 } },
    },
    workspace: {
        getConfiguration() {
            return undefined
        },
        openTextDocument: (uri: string) => ({
            getText: () => 'foo\nbar\nfoo',
            save: () => true,
        }),
        applyEdit: (edit: WorkspaceEdit) => true,
        save: () => true,
    },
    ConfigurationTarget: {
        Global: undefined,
    },
    Uri: {
        file: (path: string) => ({
            fsPath: path,
            path,
        }),
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
