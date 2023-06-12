import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { computeDiff, Diff } from './diff'
import { FixupCodeLenses } from './FixupCodeLenses'
import { ContentProvider } from './FixupContentStore'
import { FixupDecorator } from './FixupDecorator'
import { FixupDocumentEditObserver } from './FixupDocumentEditObserver'
import { FixupFile } from './FixupFile'
import { FixupFileObserver } from './FixupFileObserver'
import { FixupScheduler } from './FixupScheduler'
import { FixupTask, taskID } from './FixupTask'
import { FixupFileCollection, FixupIdleTaskRunner, FixupTextChanged } from './roles'
import { TaskViewProvider, FixupTaskTreeItem } from './TaskViewProvider'
import { CodyTaskState } from './utils'

// This class acts as the factory for Fixup Tasks and handles communication between the Tree View and editor
export class FixupController implements FixupFileCollection, FixupIdleTaskRunner, FixupTextChanged {
    private tasks = new Map<taskID, FixupTask>()
    private readonly taskViewProvider: TaskViewProvider
    private readonly files: FixupFileObserver
    private readonly editObserver: FixupDocumentEditObserver
    // TODO: Make the fixup scheduler use a cooldown timer with a longer delay
    private readonly scheduler = new FixupScheduler(10)
    private readonly decorator = new FixupDecorator()
    private readonly codelenses = new FixupCodeLenses(this)
    private readonly contentStore = new ContentProvider()

    private _disposables: vscode.Disposable[] = []

    constructor() {
        // Register commands
        this._disposables.push(
            vscode.workspace.registerTextDocumentContentProvider('cody-fixup', this.contentStore),
            vscode.commands.registerCommand('cody.fixup.open', id => this.showThisFixup(id)),
            vscode.commands.registerCommand('cody.fixup.apply', treeItem => this.applyFixups(treeItem)),
            vscode.commands.registerCommand('cody.fixup.apply-by-file', treeItem => this.applyFixups(treeItem)),
            vscode.commands.registerCommand('cody.fixup.apply-all', () => this.applyFixups()),
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
        const task = new FixupTask(fixupFile, input, editor)
        void this.setTaskState(task, CodyTaskState.asking)
        // Move the cursor to the start of the selection range to show fixup markups
        editor.selection = new vscode.Selection(task.selectionRange.start, task.selectionRange.start)
        return task.id
    }

    // Replaces content of the file before mark the task as done
    // Then update the tree view with the new task state
    public stop(taskID: taskID): void {
        const task = this.tasks.get(taskID)
        if (!task) {
            return
        }
        void this.setTaskState(task, CodyTaskState.ready)
        // show diff between current document and replacement content
        void this.contentStore.set(task.id, task.editor.document.uri)
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

    // Apply single fixup from task ID. Public for testing.
    public async apply(id: taskID): Promise<void> {
        console.log(id + ' applying')
        const task = this.tasks.get(id)
        if (!task) {
            this.discard(id)
            console.error('cannot find task')
            return
        }
        await this.applyTask(task)
    }

    // Tries to get a clean, up-to-date diff to apply. If the diff is not
    // up-to-date, it is synchronously recomputed. If the diff is not clean,
    // will return undefined. This may update the task with the newly computed
    // diff.
    private applicableDiffOrRespin(task: FixupTask, document: vscode.TextDocument): Diff | undefined {
        const bufferText = document.getText(task.selectionRange)
        let diff = task.diff
        if (task.replacement !== undefined && bufferText !== diff?.bufferText) {
            // The buffer changed since we last computed the diff.
            task.diff = diff = computeDiff(task.original, task.replacement, bufferText, task.selectionRange.start)
            this.didUpdateDiff(task)
        }
        if (!diff?.clean) {
            // TODO: Schedule a re-spin for diffs with conflicts.
            void vscode.window.showWarningMessage('applying fixup with incomplete/conflict diff is not yet implemented')
            return undefined
        }
        return diff
    }

    private async applyTask(task: FixupTask): Promise<void> {
        await this.setTaskState(task, CodyTaskState.applying)

        let editor = vscode.window.visibleTextEditors.find(editor => editor.document.uri === task.fixupFile.uri)
        if (!editor) {
            editor = await vscode.window.showTextDocument(task.fixupFile.uri)
        }

        const diff = this.applicableDiffOrRespin(task, editor.document)
        if (!diff) {
            return
        }

        editor.revealRange(task.selectionRange)
        const editOk = await editor.edit(editBuilder => {
            for (const edit of diff.edits) {
                editBuilder.replace(
                    new vscode.Range(
                        new vscode.Position(edit.range.start.line, edit.range.start.character),
                        new vscode.Position(edit.range.end.line, edit.range.end.character)
                    ),
                    edit.text
                )
            }
        })

        if (!editOk) {
            // TODO: Try to recover, for example by respinning
            void vscode.window.showWarningMessage('edit did not apply')
            return
        }

        // TODO: is this the right transition for being all done?
        // TODO: Consider keeping tasks around to resurrect them if the user
        // hits undo.
        // TODO: See if we can discard a FixupFile now.
        await this.setTaskState(task, CodyTaskState.fixed)
    }

    // Applying fixups from tree item click
    private async applyFixups(treeItem?: FixupTaskTreeItem): Promise<void> {
        // TODO: Add support for applying all fixups
        // applying fixup to all tasks
        if (!treeItem) {
            void vscode.window.showInformationMessage(
                'Applying all fixups is not implemented yet...',
                String(this.tasks.size)
            )
            return
        }
        // applying fixup to a single task
        if (treeItem.contextValue === 'task' && treeItem.id) {
            await this.apply(treeItem.id)
            return
        }
        // TODO: Add support for applying fixups from a directory
        // applying fixup to all tasks in a directory
        if (treeItem.contextValue === 'fsPath') {
            for (const task of this.tasks.values()) {
                void vscode.window.showInformationMessage(
                    'Applying fixups from a directory is not implemented yet...',
                    String(this.tasks.size)
                )
                if (task.fixupFile.uri.fsPath.endsWith(treeItem.fsPath)) {
                    return
                }
            }
            return
        }
        console.error('cannot apply fixups')
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
        this.discard(id)
        void vscode.window.showInformationMessage('Cancelling a task is not implemented yet...' + id)
        return
    }

    private discard(id: taskID): void {
        const task = this.tasks.get(id)
        if (!task) {
            return
        }
        this.needsDiffUpdate_.delete(task)
        this.codelenses.didDeleteTask(task)
        this.contentStore.delete(id)
        this.decorator.didCompleteTask(task)
        this.tasks.delete(id)
        this.taskViewProvider.removeTreeItemByID(id)
    }

    public getTaskView(): TaskViewProvider {
        return this.taskViewProvider
    }

    public getTasks(): FixupTask[] {
        return Array.from(this.tasks.values())
    }

    public async didReceiveFixupText(id: string, text: string, state: 'streaming' | 'complete'): Promise<void> {
        const task = this.tasks.get(id)
        if (!task) {
            return Promise.resolve()
        }
        if (task.state !== CodyTaskState.asking && task.state !== CodyTaskState.marking) {
            // TODO: Update this when we re-spin tasks with conflicts so that
            // we store the new text but can also display something reasonably
            // stable in the editor
            return Promise.resolve()
        }

        switch (state) {
            case 'streaming':
                task.inProgressReplacement = text
                task.marking()
                await this.setTaskState(task, CodyTaskState.marking)
                break
            case 'complete':
                task.inProgressReplacement = undefined
                task.replacement = text
                task.stop()
                await this.setTaskState(task, CodyTaskState.ready)
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

    // Handles when the range associated with a fixup task changes.
    public rangeDidChange(task: FixupTask): void {
        this.codelenses.didUpdateTask(task)
        // We don't notify the decorator about this range change; vscode
        // updates any text decorations and we can recompute them, lazily,
        // if the diff is dirtied.
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
            const task = this.needsDiffUpdate_.keys().next().value as FixupTask
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
            task.diff = computeDiff(task.original, botText, bufferText, task.selectionRange.start)
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

    // Callback function for the Fixup Task Tree View item Diff button
    private async showDiff(treeItem: FixupTaskTreeItem): Promise<void> {
        if (!treeItem?.id) {
            return
        }
        await this.diff(treeItem.id)
    }

    // Show diff between before and after edits
    private async diff(id: taskID): Promise<void> {
        const task = this.tasks.get(id)
        if (!task) {
            return
        }
        // Get an up-to-date diff
        const editor = vscode.window.visibleTextEditors.find(editor => editor.document.uri === task.fixupFile.uri)
        if (!editor) {
            return
        }
        const diff = this.applicableDiffOrRespin(task, editor.document)
        if (!diff || diff.mergedText === undefined) {
            return
        }
        // show diff view between the current document and replacement
        // Add replacement content to the temp document
        const tempDocUri = vscode.Uri.parse(`cody-fixup:${task.fixupFile.uri.fsPath}#${task.id}`)
        const doc = await vscode.workspace.openTextDocument(tempDocUri)
        const edit = new vscode.WorkspaceEdit()
        const range = task.selectionRange
        edit.replace(tempDocUri, range, diff.mergedText)
        await vscode.workspace.applyEdit(edit)
        await doc.save()

        // Show diff between current document and replacement content
        await vscode.commands.executeCommand(
            'vscode.diff',
            task.fixupFile.uri,
            tempDocUri,
            'Cody Fixup Diff View - ' + task.id,
            {
                preview: true,
                preserveFocus: false,
                selection: range,
                label: 'Cody Fixup Diff View',
                description: 'Cody Fixup Diff View: ' + task.fixupFile.uri.fsPath,
            }
        )
    }

    private setTaskState(task: FixupTask, state: CodyTaskState): Promise<FixupTask | null> {
        // TODO: Something is abusing this method to refresh code lenses. Find
        // the state 2->2 (and maybe more) callers, work out what they are
        // actually notifying, and just notify about that.
        switch (state) {
            case CodyTaskState.queued:
                task.queue()
                break
            case CodyTaskState.asking:
                task.start()
                break
            case CodyTaskState.ready:
                task.stop()
                break
            case CodyTaskState.marking:
                task.marking()
                break
            case CodyTaskState.fixed:
                task.fixed()
                break
            case CodyTaskState.error:
                task.error()
                break
            case CodyTaskState.applying:
                task.apply()
                break
        }
        if (task.state === CodyTaskState.fixed) {
            this.discard(task.id)
            return Promise.resolve(null)
        }
        // Save states of the task
        this.codelenses.didUpdateTask(task)
        // TODO: Tasks should not change IDs, so just store them once on
        // creation.
        this.tasks.set(task.id, task)
        this.taskViewProvider.setTreeItem(task)
        return Promise.resolve(task)
    }

    private reset(): void {
        this.tasks = new Map<taskID, FixupTask>()
        this.taskViewProvider.reset()
    }

    public dispose(): void {
        this.reset()
        this.codelenses.dispose()
        this.decorator.dispose()
        this.taskViewProvider.dispose()
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
