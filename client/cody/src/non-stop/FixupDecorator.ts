import * as vscode from 'vscode'

import { Diff } from './diff'

export class FixupDecorator implements vscode.Disposable {
    private decorationCodyConflictMarker_: vscode.TextEditorDecorationType
    private decorationCodyConflicted_: vscode.TextEditorDecorationType
    private decorationCodyEdited_: vscode.TextEditorDecorationType

    constructor() {
        // TODO: Switch colors depending on the theme
        this.decorationCodyConflictMarker_ = vscode.window.createTextEditorDecorationType({
            borderColor: 'orange',
            borderStyle: 'solid',
            borderWidth: '1px',
            backgroundColor: 'lightorange',
            before: {
                contentText: '\u{1F4A5}',
            },
        })
        this.decorationCodyConflicted_ = vscode.window.createTextEditorDecorationType({
            borderColor: '#FBBF77',
            backgroundColor: '#F17829',
            borderStyle: 'solid',
            borderWidth: '1px',
        })
        this.decorationCodyEdited_ = vscode.window.createTextEditorDecorationType({
            borderColor: '#9CDCFE',
            borderStyle: 'solid',
            borderWidth: '1px',
            backgroundColor: '#569CD6',
        })
    }

    public dispose(): void {
        this.decorationCodyConflictMarker_.dispose()
        this.decorationCodyConflicted_.dispose()
        this.decorationCodyEdited_.dispose()
    }

    public decorate(editor: vscode.TextEditor, diff: Diff): void {
        // TODO: Multiple fixups per file, multiple files
        if (diff.clean) {
            editor.setDecorations(this.decorationCodyConflicted_, [])
            editor.setDecorations(this.decorationCodyConflictMarker_, [])
        } else {
            editor.setDecorations(this.decorationCodyEdited_, [])
        }
        editor.setDecorations(
            diff.clean ? this.decorationCodyEdited_ : this.decorationCodyConflicted_,
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
