import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { FixupTask } from './FixupTask'
import { TaskViewProvider, FixupTaskTreeItem } from './TaskViewProvider'
import { CodyTaskState } from './utils'

type taskID = string

// This class acts as the factory for Fixup Tasks and handles communication between the Tree View and editor
export class TaskController {
    private tasks = new Map<taskID, FixupTask>()
    private readonly taskViewProvider: TaskViewProvider

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
        // Start the fixup tree view provider
        this.taskViewProvider = new TaskViewProvider()
    }

    // Adds a new task to the list of tasks
    // Then mark it as pending before sending it to the tree view for tree item creation
    public add(input: string, selection: ActiveTextEditorSelection): string | null {
        const editor = vscode.window.activeTextEditor
        if (!editor) {
            void vscode.window.showInformationMessage('No active editor found...')
            return null
        }

        // Create a task and then mark it as start
        const task = new FixupTask(input, selection, editor)
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
                filePaths.push(task.documentUri.fsPath)
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
        void vscode.window.showTextDocument(task.documentUri, { selection: task.selectionRange })
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
                if (task.documentUri.fsPath === treeItem.fsPath && task.state === CodyTaskState.done) {
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
        this.taskViewProvider.dispose()
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}
