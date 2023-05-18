import * as vscode from 'vscode'

function lineTrackerTest(change: vscode.Range, cur: vscode.Range, changeText: string): vscode.Range | null {
    if (change.start.line > cur.end.line) {
        return null
    }
    let addedLines = 0
    if (changeText.includes('\n')) {
        addedLines = changeText.split('\n').length - 1
    } else if (change.end.line - change.start.line > 0) {
        addedLines -= change.end.line - change.start.line
    }
    const newStartLine = change.start.line > cur.start.line ? cur.start.line : cur.start.line + addedLines
    const newRange = new vscode.Range(newStartLine, 0, cur.end.line + addedLines, 0)
    return newRange
}

describe('lineTracker returns a new selection range based on changes made to the current doc', () => {
    it('Return null if change occurs after the current selected range', () => {
        const cur = new vscode.Range(1, 0, 3, 0)
        const change = new vscode.Range(5, 0, 10, 0)
        const result = lineTrackerTest(change, cur, '')
        expect(result).toBeNull()
    })
    it('Update range by number of lines added within the current selected range', () => {
        const cur = new vscode.Range(5, 6, 8, 9)
        const change = new vscode.Range(6, 0, 5, 0)
        const changeText = 'line6'
        const result = lineTrackerTest(change, cur, changeText)
        expect(result).toEqual(new vscode.Range(5, 0, 8, 0))
    })
    it('Update range by number of lines removed within the current selected range', () => {
        const cur = new vscode.Range(1, 6, 5, 9)
        const change = new vscode.Range(2, 0, 3, 0)
        const changeText = 'line2\nline3'
        const result = lineTrackerTest(change, cur, changeText)
        expect(result).toEqual(new vscode.Range(1, 0, 6, 0))
    })
    it('Update range by number of lines added above the current selected range', () => {
        const cur = new vscode.Range(7, 0, 10, 0)
        const change = new vscode.Range(1, 0, 5, 0)
        const changeText = 'line1\nline2'
        const result = lineTrackerTest(change, cur, changeText)
        expect(result).toEqual(new vscode.Range(8, 0, 11, 0))
    })
    it('Update range by number of lines added overlap the current selected range', () => {
        const cur = new vscode.Range(1, 0, 3, 0)
        const change = new vscode.Range(1, 0, 3, 0)
        const changeText = 'line1\nline2\nline3'
        const result = lineTrackerTest(change, cur, changeText)
        expect(result).toEqual(new vscode.Range(3, 0, 5, 0))
    })
    it('Update range by number of lines removed before the current selected range', () => {
        const cur = new vscode.Range(5, 0, 10, 0)
        const change = new vscode.Range(1, 0, 3, 0)
        const changeText = 'line0'
        const result = lineTrackerTest(change, cur, changeText)
        expect(result).toEqual(new vscode.Range(3, 0, 8, 0))
    })
})
