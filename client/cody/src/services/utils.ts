import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'
import { SURROUNDING_LINES } from '@sourcegraph/cody-shared/src/prompt/constants'

/**
 * For tracking lines diff
 */
export async function editDocByUri(
    uri: vscode.Uri,
    lines: { start: number; end: number },
    content: string
): Promise<vscode.Range> {
    // Highlight from the start line to the length of the replacement content
    const lineDiff = content.split('\n').length - 2
    const document = await vscode.workspace.openTextDocument(uri)
    const edit = new vscode.WorkspaceEdit()
    const range = new vscode.Range(lines.start, 0, lines.end + 1, 0)
    edit.delete(document.uri, range)
    edit.insert(document.uri, new vscode.Position(lines.start, 0), content)
    await vscode.workspace.applyEdit(edit)
    return new vscode.Range(lines.start, 0, lines.start + lineDiff, 0)
}
/**
 * Get current selection from the doc uri
 * Add an extra line to the end line to prevent empty selection on single line selection
 */
export async function getFixupEditorSelection(
    docUri: vscode.Uri,
    range: vscode.Range
): Promise<{ selection: ActiveTextEditorSelection; selectionRange: vscode.Range }> {
    const activeDocument = await vscode.workspace.openTextDocument(docUri)
    const lineLength = activeDocument.lineAt(range.end.line).text.length

    const selectionRange = new vscode.Range(range.start.line, 0, range.end.line, lineLength + 1)

    const precedingText = activeDocument.getText(
        new vscode.Range(new vscode.Position(Math.max(0, range.start.line - SURROUNDING_LINES), 0), range.start)
    )
    const followingText = activeDocument.getText(
        new vscode.Range(range.end, new vscode.Position(range.end.line + 1 + SURROUNDING_LINES, 0))
    )
    // Empty selectedText will cause error in empty file
    const selection = {
        fileName: vscode.workspace.asRelativePath(docUri),
        selectedText: activeDocument.getText(selectionRange) || ' ',
        precedingText,
        followingText,
    }

    return { selection, selectionRange }
}
/**
 * Create a new range based on the current range and the updated content
 */
export function getNewRangeOnChange(
    cur: vscode.Range,
    change: { range: vscode.Range; text: string }
): vscode.Range | null {
    if (change.range.start.line > cur.end.line) {
        return null
    }
    let addedLines = 0
    if (change.text.includes('\n')) {
        addedLines = change.text.split('\n').length - 1
    } else if (change.range.end.line - change.range.start.line > 0) {
        addedLines -= change.range.end.line - change.range.start.line
    }
    const newStartLine = change.range.start.line > cur.start.line ? cur.start.line : cur.start.line + addedLines
    return new vscode.Range(newStartLine, 0, cur.end.line + addedLines, 0)
}
/**
 * Extra start and end line numbers from range
 */
export function getLineNumbersFromRange(range: vscode.Range): { start: number; end: number } {
    return { start: range.start.line, end: range.end.line }
}
/**
 * Create selection range for a single line
 * This is used for display the Cody icon and Code action on top of the first line of selected code
 */
export function singleLineRange(line: number): vscode.Range {
    return new vscode.Range(line, 0, line, 0)
}
/**
 * Generate icon path for each speaker: cody vs human (sourcegraph)
 */
export function getIconPath(speaker: string, extPath: string): vscode.Uri {
    const extensionPath = vscode.Uri.file(extPath)
    const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')
    return vscode.Uri.joinPath(webviewPath, speaker === 'cody' ? 'cody.png' : 'sourcegraph.png')
}
