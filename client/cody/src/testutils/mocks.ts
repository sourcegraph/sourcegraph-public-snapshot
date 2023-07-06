/* eslint-disable @typescript-eslint/no-empty-function */
// TODO: use implements vscode.XXX on mocked classes to ensure they match the real vscode API.
// import * as vscode from 'vscode'

import url from 'url'

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
    public isEqual(other: Range): boolean {
        return this.start.isEqual(other.start) && this.end.isEqual(other.end)
    }
}

class Uri {
    constructor(private value: string, public scheme: string, public path: string, public fsPath: string) {}

    public static parse(uri: string): Uri {
        const purl = new URL(uri)
        // NOTE: fsPath is actually way more complex than this but we have no cases
        // in our tests where its complexities are required to be implemented
        return new Uri(uri, purl.protocol, purl.pathname, purl.pathname)
    }

    public static file(pathname: string): Uri {
        return new Uri(url.pathToFileURL(pathname).toString(), 'file', pathname, pathname)
    }

    public toString() {
        return this.value
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

class EventEmitter {
    public on: () => undefined

    constructor() {
        this.on = () => undefined
    }
}

enum EndOfLine {
    /**
     * The line feed `\n` character.
     */
    LF = 1,
    /**
     * The carriage return line feed `\r\n` sequence.
     */
    CRLF = 2,
}

class CancellationTokenSource {
    public token: unknown

    constructor() {
        this.token = {
            onCancellationRequested() {},
        }
    }
}

export const vsCodeMocks = {
    Range,
    Position,
    InlineCompletionItem,
    EventEmitter,
    EndOfLine,
    CancellationTokenSource,
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
            console.error(message)
        },
        activeTextEditor: { document: { uri: Uri.file('/test.ts') }, options: { tabSize: 4 } },
        onDidChangeActiveTextEditor() {},
    },
    workspace: {
        getConfiguration() {
            return {
                get(key: string) {
                    switch (key) {
                        case 'cody.debug.filter':
                            return '.*'
                        default:
                            return ''
                    }
                },
            }
        },
        openTextDocument: (uri: string) => ({
            uri: Uri.parse(uri),
            getText: () => 'foo\nbar\nfoo',
            save: () => true,
        }),
        applyEdit: (edit: WorkspaceEdit) => true,
        save: () => true,
        getWorkspaceFolder: (uri: Uri) => ({ uri: Uri.file('/') }),
        workspaceFolders: [{ uri: Uri.file('/') }],
        asRelativePath(path: string) {
            return path
        },
        onDidChangeTextDocument() {
            return null
        },
    },
    ConfigurationTarget: {
        Global: undefined,
    },
    Uri,
    extensions: {
        getExtension() {
            return undefined
        },
    },
    InlineCompletionTriggerKind: {
        Invoke: 0,
        Automatic: 1,
    },
} as const
