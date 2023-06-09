import * as vscode from 'vscode'

import { Diff } from './diff'
import { FixupFile } from './FixupFile'
import { FixupTask } from './FixupTask'

interface TaskDecorations {
    edits: vscode.Range[]
    conflicts: vscode.Range[]
}

function makeDecorations(diff: Diff | undefined): TaskDecorations {
    if (!diff) {
        return {
            edits: [],
            conflicts: [],
        }
    }
    return {
        edits: diff.edits.map(
            edit =>
                new vscode.Range(
                    new vscode.Position(edit.range.start.line, edit.range.start.character),
                    new vscode.Position(edit.range.end.line, edit.range.end.character)
                )
        ),
        conflicts: diff.conflicts.map(
            conflict =>
                new vscode.Range(
                    new vscode.Position(conflict.start.line, conflict.start.character),
                    new vscode.Position(conflict.end.line, conflict.end.character)
                )
        ),
    }
}

// TODO: Consider constraining decorations to visible ranges.
export class FixupDecorator implements vscode.Disposable {
    private decorationCodyConflictMarker_: vscode.TextEditorDecorationType
    private decorationCodyConflicted_: vscode.TextEditorDecorationType
    private decorationCodyIncoming_: vscode.TextEditorDecorationType
    private decorations_: Map<FixupFile, Map<FixupTask, TaskDecorations>> = new Map()

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

    public didChangeVisibleTextEditors(file: FixupFile, editors: vscode.TextEditor[]): void {
        this.applyDecorations(editors, this.decorations_.get(file)?.values() || [].values())
    }

    public didUpdateDiff(task: FixupTask): void {
        this.updateTaskDecorations(task, task.diff)
    }

    // TODO: Call this so we can delete old decorations some time.
    public didCompleteTask(task: FixupTask): void {
        this.updateTaskDecorations(task, undefined)
    }

    private updateTaskDecorations(task: FixupTask, diff: Diff | undefined): void {
        const decorations = makeDecorations(diff)
        const isEmpty = decorations.edits.length === 0 && decorations.conflicts.length === 0
        let fileTasks = this.decorations_.get(task.fixupFile)
        if (!fileTasks && isEmpty) {
            // The file was not decorated; we have no decorations. Do nothing.
            return
        }
        if (isEmpty) {
            if (fileTasks?.has(task)) {
                // There were old decorations; remove them.
                fileTasks.delete(task)
                this.didChangeFileDecorations(task.fixupFile)
            }
            return
        }
        if (!fileTasks) {
            // Create the map to hold this file's decorations.
            fileTasks = new Map()
            this.decorations_.set(task.fixupFile, fileTasks)
        }
        fileTasks.set(task, decorations)
        this.didChangeFileDecorations(task.fixupFile)
    }

    private didChangeFileDecorations(file: FixupFile): void {
        // TODO: Cache the changed files and update the decorations together.
        const editors = vscode.window.visibleTextEditors.filter(editor => editor.document.uri === file.uri)
        if (!editors.length) {
            return
        }
        this.applyDecorations(editors, this.decorations_.get(file)?.values() || [].values())
    }

    private applyDecorations(editors: vscode.TextEditor[], decorations: IterableIterator<TaskDecorations>): void {
        const incoming: vscode.Range[] = []
        const conflicted: vscode.Range[] = []
        const conflicts: vscode.Range[] = []
        for (const decoration of decorations) {
            ;(decoration.conflicts.length ? conflicted : incoming).push(...decoration.edits)
            conflicts.push(...decoration.conflicts)
        }
        for (const editor of editors) {
            editor.setDecorations(this.decorationCodyConflictMarker_, conflicts)
            editor.setDecorations(this.decorationCodyConflicted_, conflicted)
            editor.setDecorations(this.decorationCodyIncoming_, incoming)
        }
    }
}
