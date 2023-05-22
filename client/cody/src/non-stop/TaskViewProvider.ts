import * as vscode from 'vscode'

import { FixupTask } from './FixupTask'
import { CodyTaskState, fixupTaskIcon, getFileNameAfterLastDash } from './utils'

export class TaskViewProvider implements vscode.TreeDataProvider<FixupTaskTreeItem> {
    /**
     * Tree items are mapped by fsPath to taskID
     */
    private treeNodes = new Map<string, FixupTaskTreeItem>()
    private treeItems = new Map<string, FixupTaskTreeItem>()

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeTreeData = new vscode.EventEmitter<FixupTaskTreeItem | undefined | void>()
    public readonly onDidChangeTreeData = this._onDidChangeTreeData.event

    constructor() {
        void vscode.commands.executeCommand('setContext', 'cody.task.view.isEmpty', true)
        this._disposables.push(vscode.commands.registerCommand('cody.task.open', taskUri => this.openFile(taskUri)))
        this._disposables.push(vscode.commands.registerCommand('cody.task.reset', () => this.reset()))
    }

    /**
     * Refresh the tree view to get the latest data
     */
    public refresh(): void {
        void vscode.commands.executeCommand('setContext', 'cody.task.view.isEmpty', this.treeNodes.size === 0)
        this._onDidChangeTreeData.fire()
    }

    /**
     * Get parents items first
     * Then returns children items for each parent item
     */
    public getChildren(element?: FixupTaskTreeItem): FixupTaskTreeItem[] {
        if (element && element.contextValue === 'fsPath') {
            return [...this.treeItems.values()].filter(item => item.fsPath === element.fsPath)
        }

        return [...this.treeNodes.values()]
    }

    /**
     * Create tree item based on provided task
     */
    public setTreeItem(task: FixupTask): void {
        const treeItem = new FixupTaskTreeItem(task.instruction, task)
        this.treeItems.set(task.id, treeItem)

        // Add fsPath to treeNodes
        const treeNode = this.treeNodes.get(task.selection.fileName) || new FixupTaskTreeItem(task.selection.fileName)
        treeNode.setChildren(task.id, task.state)
        this.treeNodes.set(task.selection.fileName, treeNode)

        this.refresh()
    }

    /**
     * Get individual tree item
     */
    public getTreeItem(element: FixupTaskTreeItem): FixupTaskTreeItem {
        return element
    }

    /**
     * Open fsPath in editor on tree item click
     */
    private openFile(uri: vscode.Uri): void {
        void vscode.window.showTextDocument(uri)
    }

    /**
     * Empty the tree view
     */
    public reset(): void {
        this.treeNodes = new Map<string, FixupTaskTreeItem>()
        this.treeItems = new Map<string, FixupTaskTreeItem>()
        this.refresh()
    }

    /**
     * Dispose the disposables
     */
    public dispose(): void {
        this.reset()
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}

class FixupTaskTreeItem extends vscode.TreeItem {
    private state: CodyTaskState = CodyTaskState.idle
    public fsPath: string

    // state for parent node
    private failed = new Set<string>()
    private tasks = new Set<string>()

    constructor(label: string, task?: FixupTask) {
        super(label)
        if (!task) {
            this.fsPath = label
            this.label = getFileNameAfterLastDash(label)
            this.contextValue = 'fsPath'
            this.collapsibleState = vscode.TreeItemCollapsibleState.Expanded
            this.description = '0 fixups'
            return
        }
        this.state = task.state
        this.id = task.id
        this.fsPath = task.selection.fileName
        this.resourceUri = task.documentUri
        this.contextValue = 'task'
        this.collapsibleState = vscode.TreeItemCollapsibleState.None
        this.tooltip = new vscode.MarkdownString(`$(zap) Task#${task.id}: ${task.instruction}`, true)
        this.command = { command: 'cody.task.open', title: 'Go to File', arguments: [task.documentUri] }

        this.updateIconPath()
    }

    // For parent node to track children states
    public setChildren(taskID: string, state: CodyTaskState): void {
        if (this.contextValue !== 'fsPath') {
            return
        }
        this.tasks.add(taskID)
        this.description = this.makeNodeDescription(state === CodyTaskState.pending)
    }

    private makeNodeDescription(running: boolean): string {
        let text = `${this.tasks.size} fixups`
        if (running) {
            text += ', 1 running'
        }
        if (this.failed.size > 0) {
            text += `, ${this.failed.size} failed`
        }
        return text
    }

    private updateIconPath(): void {
        const icon = fixupTaskIcon[this.state].icon
        const mode = fixupTaskIcon[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))
    }
}
