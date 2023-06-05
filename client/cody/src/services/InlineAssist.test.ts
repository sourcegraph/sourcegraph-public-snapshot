import * as vscode from 'vscode'

import { editDocByUri, updateRangeOnDocChange } from './InlineAssist'

describe('UpdateRangeOnDocChange returns a new selection range by calculating lines of code changed in current docs', () => {
    it('Returns current Range if change occurs after the current selected range', () => {
        const cur = new vscode.Range(1, 0, 3, 0)
        const change = new vscode.Range(5, 0, 10, 0)
        const result = updateRangeOnDocChange(cur, change, '')
        expect(result).toEqual(cur)
    })
    it('Updates range by number of lines added within the current selected range', () => {
        const cur = new vscode.Range(5, 6, 8, 9)
        const change = new vscode.Range(6, 0, 5, 0)
        const changeText = 'line6'
        const result = updateRangeOnDocChange(cur, change, changeText)
        expect(result).toEqual(new vscode.Range(5, 0, 8, 0))
    })
    it('Updates range by number of lines removed within the current selected range', () => {
        const cur = new vscode.Range(1, 6, 5, 9)
        const change = new vscode.Range(2, 0, 3, 0)
        const changeText = 'line2\nline3'
        const result = updateRangeOnDocChange(cur, change, changeText)
        expect(result).toEqual(new vscode.Range(1, 0, 6, 0))
    })
    it('Updates range by number of lines added above the current selected range', () => {
        const cur = new vscode.Range(7, 0, 10, 0)
        const change = new vscode.Range(1, 0, 5, 0)
        const changeText = 'line1\nline2'
        const result = updateRangeOnDocChange(cur, change, changeText)
        expect(result).toEqual(new vscode.Range(8, 0, 11, 0))
    })
    it('Updates range by number of lines added overlap the current selected range', () => {
        const cur = new vscode.Range(1, 0, 3, 0)
        const change = new vscode.Range(1, 0, 3, 0)
        const changeText = 'line1\nline2\nline3'
        const result = updateRangeOnDocChange(cur, change, changeText)
        expect(result).toEqual(new vscode.Range(3, 0, 5, 0))
    })
    it('Updates range by number of lines removed before the current selected range', () => {
        const cur = new vscode.Range(5, 0, 10, 0)
        const change = new vscode.Range(1, 0, 3, 0)
        const changeText = 'line0'
        const result = updateRangeOnDocChange(cur, change, changeText)
        expect(result).toEqual(new vscode.Range(3, 0, 8, 0))
    })
})

describe('editDocByUri returns a new selection range by calculating lines of code edited by Cody', () => {
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
