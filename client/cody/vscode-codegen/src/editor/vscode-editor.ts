import * as vscode from 'vscode'

import { ActiveTextEditor, ActiveTextEditorSelection, Editor } from '.'

const SURROUNDING_LINES = 50

export class VSCodeEditor implements Editor {
    public getActiveTextEditor(): ActiveTextEditor | null {
        const activeEditor = vscode.window.activeTextEditor
        const documentUri = activeEditor?.document.uri
        const documentText = activeEditor?.document.getText()
        return documentUri && documentText ? { content: documentText, filePath: documentUri.fsPath } : null
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        const activeEditor = vscode.window.activeTextEditor
        if (!activeEditor) {
            vscode.window.showErrorMessage('No code selected. Please select some code and try again.')
            return null
        }
        const selection = activeEditor.selection
        if (!selection || selection?.start.isEqual(selection.end)) {
            vscode.window.showErrorMessage('No code selected. Please select some code and try again.')
            return null
        }

        const precedingText = activeEditor.document.getText(
            new vscode.Range(
                new vscode.Position(Math.max(0, selection.start.line - SURROUNDING_LINES), 0),
                selection.start
            )
        )
        const followingText = activeEditor.document.getText(
            new vscode.Range(selection.end, new vscode.Position(selection.end.line + SURROUNDING_LINES, 0))
        )

        return {
            fileName: vscode.workspace.asRelativePath(activeEditor.document.uri.fsPath),
            selectedText: activeEditor.document.getText(selection),
            precedingText,
            followingText,
        }
    }

    public async showQuickPick(labels: string[]): Promise<string | undefined> {
        const label = await vscode.window.showQuickPick(labels)
        return label
    }

    public async showWarningMessage(message: string): Promise<void> {
        await vscode.window.showWarningMessage(message)
    }
}
