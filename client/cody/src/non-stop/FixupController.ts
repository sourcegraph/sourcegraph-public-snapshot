import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { computeDiff } from './diff'
import { FixupDecorator } from './FixupDecorator'
import { FixupDocumentEditObserver } from './FixupDocumentEditObserver'
import { FixupFile } from './FixupFile'
import { FixupFileObserver } from './FixupFileObserver'
import { FixupScheduler } from './FixupScheduler'
import { FixupTask } from './FixupTask'
import { FixupFileCollection, FixupIdleTaskRunner, FixupTextChanged } from './roles'
import { TaskViewProvider, FixupTaskTreeItem } from './TaskViewProvider'
import { CodyTaskState } from './utils'

type taskID = string

// This class acts as the factory for Fixup Tasks and handles communication between the Tree View and editor
export class FixupController implements FixupFileCollection, FixupIdleTaskRunner, FixupTextChanged {
    private tasks = new Map<taskID, FixupTask>()
    private readonly taskViewProvider: TaskViewProvider
    private readonly files: FixupFileObserver
    private readonly editObserver: FixupDocumentEditObserver
    // TODO: Make the fixup scheduler use a cooldown timer with a longer delay
    private readonly scheduler: FixupScheduler = new FixupScheduler(10)
    private readonly decorator: FixupDecorator = new FixupDecorator()

    private _disposables: vscode.Disposable[] = []

    constructor() {
        // Register commands
        this._disposables.push(vscode.commands.registerCommand('cody.fixup.open', id => this.showThisFixup(id)))
        this._disposables.push(
            vscode.commands.registerCommand('cody.fixup.apply', treeItem => this.applyFixup(treeItem))
        )
        this._disposables.push(
            vscode.commands.registerCommand('cody.fixup.apply-by-file', treeItem => this.applyDirFixups(treeItem))
        )
        this._disposables.push(
            vscode.commands.registerCommand('cody.fixup.apply-all', treeItem => this.applyAllFixups(treeItem))
        )
        this._disposables.push(vscode.commands.registerCommand('cody.fixup.diff', treeItem => this.showDiff(treeItem)))
        // Observe file renaming and deletion
        this.files = new FixupFileObserver()
        this._disposables.push(vscode.workspace.onDidRenameFiles(this.files.didRenameFiles.bind(this.files)))
        this._disposables.push(vscode.workspace.onDidDeleteFiles(this.files.didDeleteFiles.bind(this.files)))
        // Observe editor focus
        this._disposables.push(vscode.window.onDidChangeVisibleTextEditors(this.didChangeVisibleTextEditors.bind(this)))
        // Start the fixup tree view provider
        this.taskViewProvider = new TaskViewProvider()
        // Observe file edits
        this.editObserver = new FixupDocumentEditObserver(this)
        this._disposables.push(
            vscode.workspace.onDidChangeTextDocument(this.editObserver.textDocumentChanged.bind(this.editObserver))
        )
    }

    // FixupFileCollection

    public tasksForFile(file: FixupFile): FixupTask[] {
        return [...this.tasks.values()].filter(task => task.fixupFile === file)
    }

    public maybeFileForUri(uri: vscode.Uri): FixupFile | undefined {
        return this.files.maybeForUri(uri)
    }

    // FixupIdleTaskScheduler

    public scheduleIdle<T>(callback: () => T): Promise<T> {
        return this.scheduler.scheduleIdle(callback)
    }

    // Adds a new task to the list of tasks
    // Then mark it as pending before sending it to the tree view for tree item creation
    public add(input: string, selection: ActiveTextEditorSelection): string | null {
        const editor = vscode.window.activeTextEditor
        if (!editor) {
            void vscode.window.showInformationMessage('No active editor found...')
            return null
        }

        // Create a task and then mark it as started
        const fixupFile = this.files.forUri(editor.document.uri)
        const task = new FixupTask(fixupFile, input, selection, editor)
        task.start()
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', true)
        // Save states of the task
        this.tasks.set(task.id, task)
        this.taskViewProvider.setTreeItem(task)
        return task.id
    }

    // Replaces content of the file before mark the task as done
    // Then update the tree view with the new task state
    public stop(taskID: taskID): void {
        const task = this.tasks.get(taskID)
        if (!task) {
            return
        }
        task.stop()
        // Save states of the task
        this.tasks.set(task.id, task)
        this.taskViewProvider.setTreeItem(task)
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', false)
        this.getFilesWithApplicableFixups()
    }

    private getFilesWithApplicableFixups(): string[] {
        const filePaths: string[] = []
        for (const task of this.tasks.values()) {
            if (task.state === CodyTaskState.done) {
                // TODO: Handle unnamed files; grep this controller for other instances of fsPath
                filePaths.push(task.fixupFile.uri.fsPath)
            }
        }
        void vscode.commands.executeCommand('setContext', 'cody.fixup.filesWithApplicableFixups', filePaths.length > 0)
        return filePaths
    }

    // Open fsPath at the selected line in editor on tree item click
    private showThisFixup(taskID: taskID): void {
        const task = this.tasks.get(taskID)
        if (!task) {
            void vscode.window.showInformationMessage('No fixup was found...')
            return
        }
        // Create vscode Uri from task uri and selection range
        void vscode.window.showTextDocument(task.fixupFile.uri, { selection: task.selectionRange })
    }

    // TODO: Add support for applying fixup
    // Placeholder function for applying fixup
    private applyFixup(treeItem?: FixupTaskTreeItem): void {
        void vscode.window.showInformationMessage(`Applying fixup for task #${treeItem?.id} is not implemented yet...`)

        if (treeItem?.contextValue === 'task' && treeItem?.id) {
            const task = this.tasks.get(treeItem.id)
            task?.apply()
            return
        }
    }

    // TODO: Add support for applying fixup to all tasks in a directory
    // Placeholder function for applying fixup to all tasks in a directory
    private applyDirFixups(treeItem: FixupTaskTreeItem): void {
        void vscode.window.showInformationMessage('Applying fixups to a directory is not implemented yet...')

        if (treeItem?.contextValue === 'fsPath') {
            for (const task of this.tasks.values()) {
                if (task.fixupFile.uri.fsPath === treeItem.fsPath && task.state === CodyTaskState.done) {
                    task.apply()
                }
            }
        }

        this.getFilesWithApplicableFixups()
    }

    // TODO: Add support for applying all fixup
    // Placeholder function for applying fixup
    private applyAllFixups(treeItem?: FixupTaskTreeItem): void {
        void vscode.window.showInformationMessage('Applying fixup for all tasks is not implemented yet...')

        // Apply all fixups
        for (const task of this.tasks.values()) {
            if (task.state === CodyTaskState.done) {
                task.apply()
            }
        }
        // Clear task view after applying fixups
        // TODO Catch errors
        this.reset()
    }

    // TODO: Add support for showing diff
    // Placeholder function for showing diff
    private showDiff(treeItem: FixupTaskTreeItem): string {
        if (!treeItem?.id) {
            void vscode.window.showInformationMessage('No fixup was found...')
            return ''
        }
        const task = this.tasks.get(treeItem?.id)
        // TODO: Implement diff view
        void vscode.window.showInformationMessage(`Diff view for task #${task?.id} is not implemented yet...`)
        return task?.selection.selectedText || ''
    }

    public getTaskView(): TaskViewProvider {
        return this.taskViewProvider
    }

    public getTasks(): FixupTask[] {
        return Array.from(this.tasks.values())
    }

    private reset(): void {
        this.tasks = new Map<taskID, FixupTask>()
        this.taskViewProvider.reset()
    }

    /**
     * Dispose the disposables
     */
    public dispose(): void {
        this.decorator.dispose()
        this.taskViewProvider.dispose()
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }

    public didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        const task = this.tasks.get(id)
        if (!task) {
            return Promise.resolve()
        }
        if (task.state !== CodyTaskState.pending) {
            // TODO: Update this when we re-spin tasks with conflicts so that
            // we store the new text but can also display something reasonably
            // stable in the editor
            task.state = CodyTaskState.error
            return Promise.resolve()
        }

        switch (state) {
            case 'streaming':
                task.inProgressReplacement = text
                break
            case 'complete':
                task.inProgressReplacement = undefined
                task.replacement = text
                task.state = CodyTaskState.done
                break
        }

        this.textDidChange(task)
        return Promise.resolve()
    }

    // Handles changes to the source document in the fixup selection, or the
    // replacement text generated by Cody.
    public textDidChange(task: FixupTask): void {
        if (this.needsDiffUpdate_.size === 0) {
            void this.scheduler.scheduleIdle(() => this.updateDiffs())
        }
        if (!this.needsDiffUpdate_.has(task)) {
            this.needsDiffUpdate_.add(task)
        }
    }

    // Tasks where the text of the buffer, or the text provided by Cody, has
    // changed and we need to update diffs.
    private needsDiffUpdate_: Set<FixupTask> = new Set()

    // Files where the editor wasn't visible and we have delayed computing diffs
    // for tasks.
    private needsEditor_: Set<FixupFile> = new Set()

    private didChangeVisibleTextEditors(editors: readonly vscode.TextEditor[]): void {
        const editorsByFile = new Map<FixupFile, vscode.TextEditor[]>()
        for (const editor of editors) {
            const file = this.files.maybeForUri(editor.document.uri)
            if (!file) {
                continue
            }
            // Group editors by file so the decorator can apply decorations
            // in one shot.
            if (!editorsByFile.has(file)) {
                editorsByFile.set(file, [])
            }
            editorsByFile.get(file)?.push(editor)
            // If we were waiting for an editor to get text to diff against,
            // start that process now.
            if (this.needsEditor_.has(file)) {
                this.needsEditor_.delete(file)
                for (const task of this.tasksForFile(file)) {
                    if (this.needsDiffUpdate_.size === 0) {
                        void this.scheduler.scheduleIdle(() => this.updateDiffs())
                    }
                    this.needsDiffUpdate_.add(task)
                }
            }
        }
        // Apply any decorations we have to the visible editors.
        for (const [file, editors] of editorsByFile.entries()) {
            this.decorator.didChangeVisibleTextEditors(file, editors)
        }
    }

    private updateDiffs(): void {
        const deadlineMsec = Date.now() + 500

        while (this.needsDiffUpdate_.size && Date.now() < deadlineMsec) {
            const task = this.needsDiffUpdate_.keys().next().value
            this.needsDiffUpdate_.delete(task)
            const editor = vscode.window.visibleTextEditors.find(editor => editor.document.uri === task.fixupFile.uri)
            if (!editor) {
                this.needsEditor_.add(task.fixupFile)
                continue
            }
            // TODO: When Cody doesn't suggest any output something has gone
            // wrong; we should clean up. But updateDiffs also gets called to
            // process streaming output, so this isn't the place to detect or
            // recover from empty replacements.
            const botText = task.inProgressReplacement || task.replacement
            if (!botText) {
                continue
            }
            const bufferText = editor.document.getText(task.selectionRange)
            // TODO: The selection range needs to be updated for edits.
            // Switch to using a gross line-based range and updating it in the
            // FixupDocumentEditObserver.
            const diff = computeDiff(task.original, botText, bufferText, task.selectionRange.start)
            task.diff = diff
            this.didUpdateDiff(task)
        }

        if (this.needsDiffUpdate_.size) {
            // We did not get through the work; schedule more later.
            void this.scheduler.scheduleIdle(() => this.updateDiffs())
        }
    }

    private didUpdateDiff(task: FixupTask): void {
        if (!task.diff) {
            // Once we have a diff, we never go back to not having a diff.
            // If adding that transition, you must un-apply old highlights for
            // this task.
            throw new Error('unreachable')
        }
        this.decorator.didUpdateDiff(task)
        if (!task.diff.clean) {
            // TODO: If this isn't an in-progress diff, then schedule
            // a re-spin or notify failure
            return
        }
    }
}
