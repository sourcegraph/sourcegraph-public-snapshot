import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { FixupTask } from './FixupTask'
import { TaskViewProvider } from './TaskViewProvider'
import { CodyTaskState } from './utils'

type taskID = string

// This class acts as the factory for Fixup Tasks and handles communication between the Tree View and editor
export class TaskController {
    private tasks = new Map<taskID, FixupTask>()
    private readonly taskViewProvider: TaskViewProvider

    private _disposables: vscode.Disposable[] = []

    constructor() {
        this.taskViewProvider = new TaskViewProvider()
        this._disposables.push(vscode.commands.registerCommand('cody.task.open', id => this.showThisFixup(id)))
        this._disposables.push(vscode.commands.registerCommand('cody.task.apply', () => this.applyFixup()))
    }

    // Adds a new task to the list of tasks
    // Then mark it as pending before sending it to the tree view for tree item creation
    public add(input: string, selection: ActiveTextEditorSelection): string | null {
        const editor = vscode.window.activeTextEditor
        if (!editor) {
            void vscode.window.showErrorMessage('No active editor found...')
            return null
        }

        // Create a task and then mark it as start
        const task = new FixupTask(input, selection, editor)
        task.start()
        void vscode.commands.executeCommand('setContext', 'cody.task.running', true)
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
        void vscode.commands.executeCommand('setContext', 'cody.task.running', false)
    }

    // TODO: Add support for applying fixup
    // Placeholder function for applying fixup
    private applyFixup(): void {
        for (const task of this.tasks.values()) {
            if (task.state === CodyTaskState.done) {
                task.apply()
            }
        }
        // Clear task view after applying fixups
        this.reset()
    }

    // Open fsPath at the selected line in editor on tree item click
    private showThisFixup(taskID: taskID): void {
        const task = this.tasks.get(taskID)
        if (!task) {
            return
        }
        // Create vscode Uri from task uri and selection range
        void vscode.window.showTextDocument(task.documentUri, { selection: task.selectionRange })
    }

    public getTaskView(): TaskViewProvider {
        return this.taskViewProvider
    }

    private reset(): void {
        this.tasks = new Map<taskID, FixupTask>()
        this.taskViewProvider.reset()
    }
}
