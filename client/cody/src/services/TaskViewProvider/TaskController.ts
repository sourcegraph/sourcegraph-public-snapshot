import * as vscode from 'vscode'

import { CodyTaskState, CodyTask } from './CodyTask'

export class TaskController {
    public currentTaskID = ''
    private tasks: Map<string, CodyTask> = new Map<string, CodyTask>()

    // Gets a task by ID. Returns null if no task with that ID exists.
    public get(taskID: string): CodyTask | null {
        const task = this.tasks.get(taskID)
        return task || null
    }

    // Adds a new task to the queue or runs it immediately if no current task is running.
    //  Sets the task as the current running task if no current task exists. Otherwise queues the task.
    public add(task: CodyTask): void {
        const activeEditor = vscode.window.activeTextEditor
        const selection = activeEditor?.selection
        if (selection) {
            task.selectionRange = selection
        }
        // Set task as current task if tasks size is 0 with no current task
        if (!this.currentTaskID) {
            task.start()
            this.currentTaskID = task.id
        } else {
            task.queue()
        }
        this.tasks.set(task.id, task)
    }

    // Stops the currently running task. If newContent is provided, replaces the task's content with newContent.
    // Sets the next queued task as the current running task if one exists.
    public async stop(task: CodyTask, newContent: string | null): Promise<void> {
        this.currentTaskID = ''
        if (newContent) {
            await task.replace(newContent, task.getSelectionRange())
            task.stop()
        } else {
            task.error()
        }
        this.tasks.set(task.id, task)

        const nextTask = Array.from(this.tasks.values()).find(task => task.state === CodyTaskState.queued)
        if (nextTask) {
            nextTask.start()
            this.currentTaskID = nextTask.id
            this.tasks.set(nextTask.id, nextTask)
        }
    }

    public async stopByID(taskID: string, newContent: string | null): Promise<void> {
        const task = this.get(taskID)
        if (!task) {
            console.log('Task not found')
            return
        }
        await this.stop(task, newContent)
    }

    //  Return all tasks - this is used by tasks view to construct the task view
    public getTasks(): CodyTask[] {
        return Array.from(this.tasks.values())
    }

    //  Gets the currently running task. Returns null if no task is currently running.
    public getCurrentTask(): CodyTask | null {
        return this.get(this.currentTaskID)
    }

    // Deletes the task with the given ID. Stops the task first if it is currently running.
    public deleteTask(taskID: string): void {
        const task = this.tasks.get(taskID)
        if (!task) {
            return
        }
        if (task?.state === CodyTaskState.pending) {
            task.stop()
            return this.deleteTask(taskID)
        }
        this.tasks.delete(taskID)
    }

    public deleteAll(): void {
        this.tasks = new Map()
    }
}
