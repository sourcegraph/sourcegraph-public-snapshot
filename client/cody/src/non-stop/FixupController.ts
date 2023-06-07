import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { computeDiff } from './diff'
import { FixupCodeLenses } from './FixupCodeLenses'
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
    private readonly codelenses: FixupCodeLenses = new FixupCodeLenses()

    private _disposables: vscode.Disposable[] = []

    constructor() {
        // Register commands
        this._disposables.push(
            vscode.commands.registerCommand('cody.fixup.open', id => this.showThisFixup(id)),
            vscode.commands.registerCommand('cody.fixup.apply', treeItem => this.applyFixup(treeItem)),
            vscode.commands.registerCommand('cody.fixup.apply-by-file', treeItem => this.applyDirFixups(treeItem)),
            vscode.commands.registerCommand('cody.fixup.apply-all', () => this.applyAllFixups()),
            vscode.commands.registerCommand('cody.fixup.diff', treeItem => this.showDiff(treeItem)),
            vscode.commands.registerCommand('cody.fixup.codelens.apply', id => this.apply(id)),
            vscode.commands.registerCommand('cody.fixup.codelens.diff', id => this.diff(id)),
            vscode.commands.registerCommand('cody.fixup.codelens.discard', id => this.discard(id)),
            vscode.commands.registerCommand('cody.fixup.codelens.edit', id => this.edit(id)),
            vscode.commands.registerCommand('cody.fixup.codelens.cancel', id => this.cancel(id))
        )
        // Observe file renaming and deletion
        this.files = new FixupFileObserver()
        this._disposables.push(vscode.workspace.onDidRenameFiles(this.files.didRenameFiles.bind(this.files)))
        this._disposables.push(vscode.workspace.onDidDeleteFiles(this.files.didDeleteFiles.bind(this.files)))
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
        this.codelenses.set(task.id, CodyTaskState.queued, task.selectionRange)
        task.start()
        this.codelenses.set(task.id, CodyTaskState.pending, task.selectionRange)
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
        this.codelenses.set(task.id, task.state, task.selectionRange)
        // Save states of the task
        this.tasks.set(task.id, task)
        this.taskViewProvider.setTreeItem(task)
        void vscode.commands.executeCommand('setContext', 'cody.fixup.running', false)
        this.getFilesWithApplicableFixups()
    }

    private getFilesWithApplicableFixups(): string[] {
        const filePaths: string[] = []
        for (const task of this.tasks.values()) {
            // Remove fixed tasks
            if (task.state === CodyTaskState.fixed) {
                this.discard(task.id)
                continue
            }
            if (task.state === CodyTaskState.done) {
                // TODO: Handle unnamed files; grep this controller for other instances of fsPath
                filePaths.push(task.fixupFile.uri.fsPath)
            }
            // Update CodeLenses
            this.codelenses.set(task.id, task.state, task.selectionRange)
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

    // Apply single fixup from task ID
    private async apply(id: taskID): Promise<void> {
        await this.applyFixupById(id)
        this.getFilesWithApplicableFixups()
    }

    // Applying single fixup from tree item
    private async applyFixup(treeItem?: FixupTaskTreeItem): Promise<void> {
        if (treeItem?.contextValue !== 'task' || !treeItem?.id) {
            return
        }
        await this.applyFixupById(treeItem.id)
    }

    // Apply fixup to a task by id
    private async applyFixupById(id: taskID): Promise<void> {
        const task = this.tasks.get(id)
        if (!task) {
            return
        }
        this.codelenses.set(task.id, CodyTaskState.applying, task.selectionRange)
        await task.apply()
        this.tasks.set(task.id, task)
        if (task.state === CodyTaskState.fixed) {
            this.discard(task.id)
        }
    }

    // Placeholder function for applying fixup to all tasks in a directory
    private async applyDirFixups(treeItem: FixupTaskTreeItem): Promise<void> {
        if (treeItem?.contextValue === 'fsPath') {
            for (const task of this.tasks.values()) {
                if (task.fixupFile.uri.fsPath === treeItem.fsPath && task.state === CodyTaskState.done) {
                    await this.applyFixupById(task.id)
                }
            }
            this.getFilesWithApplicableFixups()
        }
    }

    // Placeholder function for applying fixup
    private async applyAllFixups(): Promise<void> {
        void vscode.window.showInformationMessage('Applying fixup for all tasks is not implemented yet...')
        // Apply all fixups
        for (const task of this.tasks.values()) {
            if (task.state === CodyTaskState.done) {
                await this.applyFixupById(task.id)
            }
        }

        this.getFilesWithApplicableFixups()
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

    // TODO: Add support for showing diff by task id
    private async diff(id: taskID): Promise<void> {
        const task = this.tasks.get(id)
        // show quick diff view between the current document and replacement
        const origin = task?.selection.selectedText
        const replacement = task?.replacement
        if (!origin || !replacement) {
            return
        }
        // show diff between current document and replacement content
        const tempDocUri = vscode.Uri.parse('untitled:' + task.fixupFile.uri.fsPath)
        // Add replacement content to the temp document
        const edit = new vscode.WorkspaceEdit()
        edit.replace(tempDocUri, task.selectionRange, replacement)
        await vscode.workspace.applyEdit(edit)
        // Show diff between current document and replacement content
        await vscode.commands.executeCommand(
            'vscode.diff',
            tempDocUri,
            task.fixupFile.uri,
            'Fixup: ' + task.fixupFile.uri.fsPath,
            {
                preview: true,
                selection: task.selectionRange,
                label: 'Fixup: ' + task.fixupFile.uri.fsPath,
                description: 'Fixup: ' + task.fixupFile.uri.fsPath,
            }
        )
    }

    // TODO: Add support for editing a fixup task
    // Placeholder function for editing a fixup task
    private edit(id: taskID): void {
        void vscode.window.showInformationMessage('Editing a task is not implemented yet...' + id)
        return
    }

    // TODO: Add support for cancelling a fixup task
    // Placeholder function for cancelling a fixup task
    private cancel(id: taskID): void {
        void vscode.window.showInformationMessage('Cancelling a task is not implemented yet...' + id)
        return
    }

    private discard(id: taskID): void {
        this.tasks.delete(id)
        this.taskViewProvider.removeTreeItemByID(id)
        this.codelenses.remove(id)
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
        this.reset()
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
            return Promise.resolve()
        }

        switch (state) {
            case 'streaming':
                task.inProgressReplacement = text
                break
            case 'complete':
                task.inProgressReplacement = undefined
                task.replacement = text
                task.stop()
                this.codelenses.set(task.id, task.state, task.selectionRange)
                break
        }

        this.textDidChange(task)
        return Promise.resolve()
    }

    // Handles changes to the source document in the fixup selection, or the
    // replacement text generated by Cody.
    public textDidChange(task: FixupTask): void {
        if (task.state === CodyTaskState.fixed) {
            this.needsDiffUpdate_.delete(task)
        }
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

    // TODO: Hook up onDidChangeVisibleTextEditors to apply decorations to
    // editors as they become visible

    // TODO: Move the core of this method to a separate component
    private updateDiffs(): void {
        const deadlineMsec = Date.now() + 500

        while (this.needsDiffUpdate_.size && Date.now() < deadlineMsec) {
            const task = this.needsDiffUpdate_.keys().next().value as FixupTask
            this.needsDiffUpdate_.delete(task)
            const editor = vscode.window.visibleTextEditors.find(editor => editor.document.uri === task.fixupFile.uri)
            if (!editor) {
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
            // TODO: Cache the source text and diff edits for application
            console.log(botText)
            // TODO: Cache the diff output on the fixup so it can be applied;
            // have the decorator reapply decorations to visible editors
            this.decorator.decorate(editor, diff)
            this.codelenses.set(task.id, task.state, task.selectionRange)
            if (!diff.clean) {
                // TODO: If this isn't an in-progress diff, then schedule
                // a re-spin or notify failure
                continue
            }
        }
    }
}
