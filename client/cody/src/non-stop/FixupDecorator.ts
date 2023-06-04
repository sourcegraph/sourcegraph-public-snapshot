import * as vscode from 'vscode'

import { Diff } from './diff'

export class FixupDecorator implements vscode.Disposable {
    private decorationCodyConflictMarker_: vscode.TextEditorDecorationType
    private decorationCodyConflicted_: vscode.TextEditorDecorationType
    private decorationCodyIncoming_: vscode.TextEditorDecorationType

    constructor() {
        this.decorationCodyConflictMarker_ = vscode.window.createTextEditorDecorationType({
            backgroundColor: new vscode.ThemeColor('cody.fixup.conflictBackground'),
            borderColor: new vscode.ThemeColor('cody.fixup.conflictBorder'),
            borderStyle: 'solid',
            borderWidth: '1px',
        })
        this.decorationCodyConflicted_ = vscode.window.createTextEditorDecorationType({
            backgroundColor: new vscode.ThemeColor('cody.fixup.conflictedBackground'),
            borderColor: new vscode.ThemeColor('cody.fixup.conflictedBorder'),
            borderStyle: 'solid',
            borderWidth: '1px',
        })
        this.decorationCodyIncoming_ = vscode.window.createTextEditorDecorationType({
            backgroundColor: new vscode.ThemeColor('cody.fixup.incomingBackground'),
            borderColor: new vscode.ThemeColor('cody.fixup.incomingBorder'),
            borderStyle: 'solid',
            borderWidth: '1px',
        })
    }

    public dispose(): void {
        this.decorationCodyConflictMarker_.dispose()
        this.decorationCodyConflicted_.dispose()
        this.decorationCodyIncoming_.dispose()
    }

    public decorate(editor: vscode.TextEditor, diff: Diff): void {
        // TODO: Multiple fixups per file, multiple files
        if (diff.clean) {
            editor.setDecorations(this.decorationCodyConflicted_, [])
            editor.setDecorations(this.decorationCodyConflictMarker_, [])
        } else {
            editor.setDecorations(this.decorationCodyIncoming_, [])
        }
        editor.setDecorations(
            diff.clean ? this.decorationCodyIncoming_ : this.decorationCodyConflicted_,
            diff.edits.map(
                edit =>
                    new vscode.Range(
                        new vscode.Position(edit.range.start.line, edit.range.start.character),
                        new vscode.Position(edit.range.end.line, edit.range.end.character)
                    )
            )
        )
        editor.setDecorations(
            this.decorationCodyConflictMarker_,
            diff.conflicts.map(
                conflict =>
                    new vscode.Range(
                        new vscode.Position(conflict.start.line, conflict.start.character),
                        new vscode.Position(conflict.end.line, conflict.end.character)
                    )
            )
        )
    }
}
