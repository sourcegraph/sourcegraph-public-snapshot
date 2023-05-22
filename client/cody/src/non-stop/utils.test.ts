import * as vscode from 'vscode'

import { getFileNameAfterLastDash, editDocByUri } from './utils'

jest.unmock('./utils')

jest.mock('vscode', () => {
    class Position {
        public line: number
        public character: number

        constructor(line: number, character: number) {
            this.line = line
            this.character = character
        }
    }
    class Range {
        public startLine: number
        public startCharacter: number
        public endLine: number
        public endCharacter: number

        constructor(startLine: number, startCharacter: number, endLine: number, endCharacter: number) {
            this.startLine = startLine
            this.startCharacter = startCharacter
            this.endLine = endLine
            this.endCharacter = endCharacter
        }
    }
    // TODO: Implement delete and insert mocks
    class WorkspaceEdit {
        public delete(uri: vscode.Uri, range: Range): void {}
        public insert(uri: vscode.Uri, position: Position, content: string): void {}
    }

    return {
        Position,
        Range,
        WorkspaceEdit,
        Uri: {
            file: (path: string) => path,
        },
        workspace: {
            openTextDocument: (uri: string) => ({
                getText: () => 'foo\nbar\nfoo',
                save: () => true,
            }),
            applyEdit: (edit: WorkspaceEdit) => true,
            save: () => true,
        },
        window: {
            showTextDocument: (uri: string) => ({
                edit: (callback: (editBuilder: any) => void) => {
                    const editBuilder = {
                        replace: (range: Range, content: string) => range,
                    }
                    callback(editBuilder)
                },
            }),
        },
    }
})

describe('editDocByUri', () => {
    test('replaces a single line in a document', async () => {
        const uri = vscode.Uri.file('/tmp/test.txt')
        const lines = { start: 1, end: 3 }
        const content = 'foo\nfoo\nfoo'
        const range = await editDocByUri(uri, lines, content)
        expect(range).toEqual(new vscode.Range(1, 0, 2, 0))
    })

    test('replaces multiple lines in a document', async () => {
        const uri = vscode.Uri.file('/tmp/test.txt')
        const lines = { start: 1, end: 3 }
        const content = 'foo\nbar\nfoo\nbar\nfoo'
        const range = await editDocByUri(uri, lines, content)
        expect(range).toEqual(new vscode.Range(1, 0, 4, 0))
    })
})

// Test for getFileNameAfterLastDash
describe('getFileNameAfterLastDash', () => {
    test('gets the last part of the file path after the last slash', () => {
        const filePath = '/path/to/file.txt'
        const fileName = 'file.txt'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
    test('get file name when there is no slash', () => {
        const filePath = 'file.txt'
        const fileName = 'file.txt'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
    test('get file name when there is no extension', () => {
        const filePath = 'file'
        const fileName = 'file'
        expect(getFileNameAfterLastDash(filePath)).toEqual(fileName)
    })
})
