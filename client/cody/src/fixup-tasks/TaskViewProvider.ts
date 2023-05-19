import * as vscode from 'vscode'

import { ActiveTextEditorSelection } from '@sourcegraph/cody-shared/src/editor'

import { FixupTask } from './FixupTask'

export class TaskViewProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    /**
     * FixupTask objects mapped by taskID
     */
    private tasks = new Map<string, FixupTask>()
    /**
     * Tree items mapped by:
     * fsPath = parent items
     * taskID = child items
     */
    private treeItems = new Map<string, vscode.TreeItem>()

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeTreeData = new vscode.EventEmitter<vscode.TreeItem | undefined | void>()
    public readonly onDidChangeTreeData = this._onDidChangeTreeData.event

    constructor() {
        this._disposables.push(vscode.commands.registerCommand('cody.task.open', taskUri => this.openFile(taskUri)))
    }
    /**
     * Refresh the tree view to get the latest data
     */
    public refresh(): void {
        this._onDidChangeTreeData.fire()
    }
    /**
     * Get parents items first
     * Then returns children items for each parent item
     */
    public getChildren(element?: vscode.TreeItem): vscode.TreeItem[] {
        if (element && element.contextValue === 'fsPath') {
            return [...this.treeItems.values()].filter(
                item => item.contextValue === 'task' && item.resourceUri?.fsPath === element.resourceUri?.fsPath
            )
        }
        return [...this.treeItems.values()].filter(item => item.contextValue === 'fsPath')
    }
    /**
     * Get individual tree item
     */
    public getTreeItem(element: vscode.TreeItem): vscode.TreeItem {
        return element
    }
    /**
     * Create a new tree item for a task and add it to the task list that is used to create the tree view
     */
    public newTask(taskID: string, input: string, selection: ActiveTextEditorSelection): void {
        const editor = vscode.window.activeTextEditor
        if (!editor) {
            return
        }
        // Create new task object
        const newTask = new FixupTask(taskID, input, selection, editor)

        // Create new tree items - fsPath
        const newParentItem = new vscode.TreeItem(selection.fileName)
        newParentItem.id = selection.fileName
        newParentItem.resourceUri = newTask.documentUri
        newParentItem.contextValue = 'fsPath'
        newParentItem.collapsibleState = vscode.TreeItemCollapsibleState.Expanded
        this.treeItems.set(selection.fileName, newParentItem)

        // Create new child item - task under each fsPath
        const newChildItem = new vscode.TreeItem(newTask.instruction)
        newChildItem.contextValue = 'task'
        newChildItem.description = taskID
        newChildItem.iconPath = newTask.iconPath
        newChildItem.id = taskID
        newChildItem.resourceUri = newTask.documentUri
        newChildItem.tooltip = new vscode.MarkdownString(`$(zap) Task#${taskID}: ${newTask.instruction}`, true)
        newChildItem.command = { command: 'cody.task.open', title: 'Check Task', arguments: [newTask.documentUri] }
        this.treeItems.set(taskID, newChildItem)

        // Start and set task
        newTask.start()
        this.tasks.set(taskID, newTask)
        this.refresh()
    }
    /**
     * Mark task as completed and start replacement in doc
     * Log error in output if no replacement is provided
     */
    public async stopTask(taskID: string, content: string | null): Promise<void> {
        const task = this.tasks.get(taskID)
        const treeItem = this.treeItems.get(taskID)
        if (!task || !treeItem) {
            console.error('Task not found: ' + taskID)
            return
        }
        if (!content) {
            task.error('Cody did not provide any replacement for Tash#' + taskID)
            return
        }
        await task.replace(content, task.getSelectionRange())
        // Update tree item
        this.tasks.set(taskID, task)
        treeItem.iconPath = task.iconPath
        this.treeItems.set(taskID, treeItem)
        this.refresh()
    }
    /**
     * Open fsPath in editor on tree item click
     */
    private openFile(uri: vscode.Uri): void {
        void vscode.window.showTextDocument(uri)
    }
    /**
     * Dispose the disposables
     */
    public dispose(): void {
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this.tasks = new Map<string, FixupTask>()
        this.treeItems = new Map<string, vscode.TreeItem>()
        this._disposables = []
    }
}
