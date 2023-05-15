import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { CodyTask } from './CodyTask'
import { TaskController } from './TaskController'

export class TaskViewProvider implements vscode.TreeDataProvider<CodyTask> {
    // List of tasks to create the tree view
    public tasks: CodyTask[] = []

    // TaskController handles managing Cody tasks.
    public controller: TaskController

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeTreeData: vscode.EventEmitter<CodyTask | undefined | void> = new vscode.EventEmitter<
        CodyTask | undefined | void
    >()
    public readonly onDidChangeTreeData: vscode.Event<CodyTask | undefined | void> = this._onDidChangeTreeData.event

    constructor() {
        this.controller = new TaskController()
    }

    public getTreeData(): CodyTask[] {
        return this.controller.getTasks().reverse()
    }

    public refresh(): void {
        const newTasks = this.getTreeData()
        this.tasks = newTasks
        this._onDidChangeTreeData.fire()
    }

    public getTreeItem(element: CodyTask): vscode.TreeItem {
        const tooltip = new vscode.MarkdownString(`$(zap) ${element.uri}`, true)
        element.tooltip = tooltip
        return element
    }

    public getChildren(element?: CodyTask): CodyTask[] {
        const children = this.tasks || element || []
        return children
    }

    // Create a new task and add it to task list
    public newTask(taskID: string, input: string, selection: ActiveTextEditorSelection): void {
        const task = new CodyTask(taskID, input, selection)
        this.controller.add(task)
        this.refresh()
    }

    // Mark task as completed and start replacement in doc
    public async stopTask(taskID: string, content?: string): Promise<void> {
        await this.controller.stopByID(taskID, content)
        this.refresh()
    }

    /**
     * Dispose the disposables
     */
    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this.tasks = []
        this._disposables = []
    }
}
