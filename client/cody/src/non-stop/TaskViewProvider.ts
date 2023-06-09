import * as vscode from 'vscode'

import { FixupTask } from './FixupTask'
import { CodyTaskState, fixupTaskList, getFileNameAfterLastDash } from './utils'

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
        void vscode.commands.executeCommand('setContext', 'cody.fixup.view.isEmpty', true)
    }

    /**
     * Refresh the tree view to get the latest data
     */
    public refresh(): void {
        void vscode.commands.executeCommand('setContext', 'cody.fixup.view.isEmpty', this.treeNodes.size === 0)
        this._onDidChangeTreeData.fire()
    }

    /**
     * Get parents items first
     * Then returns children items for each parent item
     */
    public getChildren(element?: FixupTaskTreeItem): FixupTaskTreeItem[] {
        if (element && element.contextValue === 'fsPath') {
            const tasksByFsPath = [...this.treeItems.values()].filter(item => item.fsPath === element.fsPath)
            if (tasksByFsPath.length === 0) {
                this.treeNodes.delete(element.fsPath)
                return []
            }
            return tasksByFsPath
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
        const treeNode = this.treeNodes.get(task.fixupFile.fileName) || new FixupTaskTreeItem(task.fixupFile.fileName)
        treeNode.addChildren(task.id, task.state)
        this.treeNodes.set(task.fixupFile.fileName, treeNode)

        this.refresh()
    }

    /**
     * Get individual tree item
     */
    public getTreeItem(element: FixupTaskTreeItem): FixupTaskTreeItem {
        return element
    }

    public removeTreeItemByID(taskID: taskID): void {
        const task = this.treeItems.get(taskID)
        if (!task) {
            return
        }
        this.treeItems.delete(taskID)
        const treeNode = this.treeNodes.get(task.fsPath)
        treeNode?.removeChild(taskID)
        if (treeNode && treeNode.tasks.size === 0) {
            this.treeNodes.delete(task.fsPath)
        }
        this.refresh()
    }

    public removeTreeItemsByFileName(fileName: fileName): void {
        const tasks = this.treeNodes.get(fileName)
        if (!tasks) {
            return
        }
        for (const task of tasks.tasks) {
            this.treeItems.delete(task)
        }
        this.treeNodes.delete(fileName)
        this.refresh()
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

export class FixupTaskTreeItem extends vscode.TreeItem {
    private state: CodyTaskState = CodyTaskState.idle
    public fsPath: string

    // state for parent node
    private failed = new Set<string>()
    public tasks = new Set<string>()

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
        this.fsPath = task.fixupFile.fileName
        // TODO: Files change URIs when they are renamed, so add a change
        // notification and don't cache this here.
        this.resourceUri = task.fixupFile.uri
        this.contextValue = 'task'
        this.collapsibleState = vscode.TreeItemCollapsibleState.None
        this.tooltip = new vscode.MarkdownString(`Task #${task.id}: ${task.instruction}`, true)
        this.command = { command: 'cody.fixup.open', title: 'Go to File', arguments: [task.id] }

        this.updateIconPath()
    }

    // For parent node to track children states
    public addChildren(taskID: string, state: CodyTaskState): void {
        if (this.contextValue !== 'fsPath' || state === CodyTaskState.fixed) {
            return
        }
        this.tasks.add(taskID)
        this.description = this.makeNodeDescription(state)
    }

    public removeChild(taskID: string): void {
        this.tasks.delete(taskID)
    }

    private makeNodeDescription(state: CodyTaskState): string {
        const tasksSize = this.tasks.size
        const failedSize = this.failed.size
        let text = `${tasksSize} ${tasksSize > 1 ? 'fixups' : 'fixup'}`
        let ready = tasksSize - failedSize

        switch (state) {
            case CodyTaskState.asking:
                text += ', 1 running'
                ready--
                break
            case CodyTaskState.applying:
                text += ', 1 applying'
                ready--
                break
            case CodyTaskState.fixed:
                ready--
                break
        }
        if (failedSize > 0) {
            text += `, ${failedSize} failed`
        }
        if (ready > 0) {
            text += `, ${ready} ready`
        }
        void vscode.commands.executeCommand('setContext', 'cody.fixup.filesWithApplicableFixups', ready < 1)
        return text
    }

    private updateIconPath(): void {
        const icon = fixupTaskList[this.state].icon
        const mode = fixupTaskList[this.state].id
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(mode))
    }
}
