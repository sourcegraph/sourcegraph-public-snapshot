import * as vscode from 'vscode'

import { FixupTask } from './FixupTask'
import { CodyTaskState, fixupTaskIcon, getFileNameAfterLastDash } from './utils'

type taskID = string
type fileName = string
export class TaskViewProvider implements vscode.TreeDataProvider<FixupTaskTreeItem> {
    /**
     * Tree items are mapped by fsPath to taskID
     */
    // Add type alias for Map key
    private treeNodes = new Map<fileName, FixupTaskTreeItem>()
    private treeItems = new Map<taskID, FixupTaskTreeItem>()

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeTreeData = new vscode.EventEmitter<FixupTaskTreeItem | undefined | void>()
    public readonly onDidChangeTreeData = this._onDidChangeTreeData.event

    constructor() {
        void vscode.commands.executeCommand('setContext', 'cody.task.view.isEmpty', true)
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
        treeNode.addChildren(task.id, task.state)
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
     * Empty the tree view
     */
    public reset(): void {
        this.treeNodes = new Map<fileName, FixupTaskTreeItem>()
        this.treeItems = new Map<taskID, FixupTaskTreeItem>()
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
            this.tooltip = label
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
        this.tooltip = new vscode.MarkdownString(`Task #${task.id}: ${task.instruction}`, true)
        this.command = { command: 'cody.task.open', title: 'Go to File', arguments: [task.id] }

        this.updateIconPath()
    }

    // For parent node to track children states
    public addChildren(taskID: string, state: CodyTaskState): void {
        if (this.contextValue !== 'fsPath') {
            return
        }
        this.tasks.add(taskID)
        this.description = this.makeNodeDescription(state)
    }

    private makeNodeDescription(state: CodyTaskState): string {
        let text = `${this.tasks.size} fixups`
        let ready = this.tasks.size - this.failed.size
        if (state === CodyTaskState.pending) {
            text += ', 1 running'
            ready--
        } else if (state === CodyTaskState.applying) {
            text += ', 1 applying'
            ready--
        }
        if (this.failed.size > 0) {
            text += `, ${this.failed.size} failed`
        }
        if (ready > 0) {
            text += `, ${ready} ready`
        }
        return text
    }

    private updateIconPath(): void {
        const icon = fixupTaskIcon[this.state].icon
        const mode = fixupTaskIcon[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))
    }
}
