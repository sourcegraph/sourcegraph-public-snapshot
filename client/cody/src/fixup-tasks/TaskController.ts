import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { FixupTask } from './FixupTask'
import { TaskViewProvider } from './TaskViewProvider'

// This class acts as the factory for Fixup Tasks and handles communication between the Tree View and editor
export class TaskController {
    private tasks = new Map<string, FixupTask>()
    private readonly taskViewProvider: TaskViewProvider

    constructor() {
        this.taskViewProvider = new TaskViewProvider()
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
    public async stop(taskID: string, content: string | null): Promise<void> {
        const task = this.tasks.get(taskID)
        if (!task) {
            return
        }
        // Runs replacement
        await task.replace(content, task.getSelectionRange())
        // Save states of the task
        this.tasks.set(task.id, task)
        this.taskViewProvider.setTreeItem(task)
        void vscode.commands.executeCommand('setContext', 'cody.task.running', false)
    }

    public getTaskView(): TaskViewProvider {
        return this.taskViewProvider
    }

    public reset(): void {
        this.tasks = new Map<string, FixupTask>()
        this.taskViewProvider.reset()
    }
}
