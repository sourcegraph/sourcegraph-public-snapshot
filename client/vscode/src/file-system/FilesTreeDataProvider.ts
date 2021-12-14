import * as vscode from 'vscode'

import { log } from '../log'

import { SourcegraphFileSystemProvider } from './SourcegraphFileSystemProvider'
import { SourcegraphUri } from './SourcegraphUri'

export class FilesTreeDataProvider implements vscode.TreeDataProvider<string> {
    constructor(public readonly fs: SourcegraphFileSystemProvider) {
        fs.onDidDownloadRepositoryFilenames(() => this.didChangeTreeData.fire(undefined))
    }

    private _isViewVisible = false
    private isExpandedNode = new Set<string>()
    private treeView: vscode.TreeView<string> | undefined
    private activeUri: vscode.Uri | undefined
    private didFocusToken = new vscode.CancellationTokenSource()
    private treeItemCache = new Map<string, vscode.TreeItem>()
    private readonly didChangeTreeData = new vscode.EventEmitter<string | undefined>()
    public readonly onDidChangeTreeData: vscode.Event<string | undefined> = this.didChangeTreeData.event

    public activeTextDocument(): SourcegraphUri | undefined {
        return this.activeUri && this.activeUri.scheme === 'sourcegraph'
            ? this.fs.sourcegraphUri(this.activeUri)
            : undefined
    }
    public isViewVisible(): boolean {
        return this._isViewVisible
    }
    public setTreeView(treeView: vscode.TreeView<string>): void {
        this.treeView = treeView
        treeView.onDidChangeVisibility(event => {
            const didBecomeVisible = !this._isViewVisible && event.visible
            this._isViewVisible = event.visible
            if (didBecomeVisible) {
                // NOTE: do not remove the line below even if you think it
                // doesn't have an effect. Before you remove this line, make
                // sure that the following steps don't cause the "Collapse All"
                // button to become disabled:
                //   1. Close "Files" view.
                //   2. Execute "Reload window" command.
                //   3. After VS Code loads, open the "Files" view.
                this.didChangeTreeData.fire(undefined)

                this.didFocus(this.activeUri).then(
                    () => {},
                    () => {}
                )
            }
        })
        treeView.onDidExpandElement(event => {
            this.isExpandedNode.add(event.element)
        })
        treeView.onDidCollapseElement(event => {
            this.isExpandedNode.delete(event.element)
        })
    }

    public async getParent(uriString?: string): Promise<string | undefined> {
        // log.appendLine(`getParent(${uriString})`)
        try {
            // Implementation note: this method is not implemented as
            // `SourcegraphUri.parse(uri).parentUri()` because that would return
            // URIs to directories that don't exist because they have no siblings
            // and are therefore automatically merged with their parent. For example,
            // imagine the following folder structure:
            //   .gitignore
            //   .github/workflows/ci.yml
            //   src/command.ts
            //   src/browse.ts
            // The parent of `.github/workflows/ci.yml` is `.github/` because the `workflows/`
            // directory has no sibling.
            if (!uriString) {
                return undefined
            }
            const uri = SourcegraphUri.parse(uriString)
            if (!uri.path) {
                return undefined
            }
            let ancestor: string | undefined = uri.repositoryUri()
            let children = await this.getChildren(ancestor)
            while (ancestor) {
                const isParent = children?.includes(uriString)
                if (isParent) {
                    break
                }
                ancestor = children?.find(childUri => {
                    const child = SourcegraphUri.parse(childUri)
                    return child.path && uri.path?.startsWith(child.path + '/')
                })
                if (!ancestor) {
                    log.errorAndThrow(`getParent(${uriString || 'undefined'}) nothing startsWith`)
                }
                children = await this.getChildren(ancestor)
            }
            return ancestor
        } catch (error) {
            log.errorAndThrow(`getParent(${uriString || 'undefined'})`, error)
            return undefined
        }
    }

    public async getChildren(uriString?: string): Promise<string[] | undefined> {
        // log.appendLine(`getChildren(${uriString})`)
        try {
            if (!uriString) {
                // await this.getTreeItem(undefined)
                const repos = [...this.fs.allRepositoryUris()]
                return repos.map(repo => repo.replace('https://', 'sourcegraph://'))
            }
            const uri = SourcegraphUri.parse(uriString)
            const tree = await this.fs.getFileTree(uri)
            const directChildren = tree.directChildren(uri.path || '')
            for (const child of directChildren) {
                this.treeItemCache.set(child, this.newTreeItem(SourcegraphUri.parse(child), uri, directChildren.length))
            }
            return directChildren
        } catch (error) {
            return log.errorAndThrow(`getChildren(${uriString || ''})`, error)
        }
    }

    public async focusActiveFile(): Promise<void> {
        await vscode.commands.executeCommand('sourcegraph.files.focus')
        await this.didFocus(this.activeUri)
    }

    public async didFocus(vscodeUri: vscode.Uri | undefined): Promise<void> {
        log.appendLine(`didFocus=${vscodeUri?.toString(true) || 'undefined'}`)
        this.didFocusToken.cancel()
        this.didFocusToken = new vscode.CancellationTokenSource()
        this.activeUri = vscodeUri
        vscode.commands
            .executeCommand('setContext', 'sourcegraph.canFocusActiveDocument', vscodeUri?.scheme === 'sourcegraph')
            .then(
                () => {},
                () => {}
            )
        if (vscodeUri && vscodeUri.scheme === 'sourcegraph' && this.treeView && this._isViewVisible) {
            const uri = this.fs.sourcegraphUri(vscodeUri)
            if (uri.uri === this.fs.emptyFileUri()) {
                return
            }
            await this.fs.downloadFiles(uri)
            await this.didFocusString(uri, true, this.didFocusToken.token)
        }
    }

    public async getTreeItem(uriString: string): Promise<vscode.TreeItem> {
        // log.appendLine(`getTreeItem(${uriString})`)
        try {
            const fromCache = this.treeItemCache.get(uriString)
            if (fromCache) {
                return fromCache
            }
            const uri = SourcegraphUri.parse(uriString)
            const parentUri = await this.getParent(uri.uri)
            return this.newTreeItem(uri, parentUri ? SourcegraphUri.parse(parentUri) : undefined, 0)
        } catch (error) {
            log.errorAndThrow(`getTreeItem(${uriString})`, error)
        }
        return {}
    }

    private async didFocusString(
        uri: SourcegraphUri,
        isDestinationNode: boolean,
        token: vscode.CancellationToken
    ): Promise<void> {
        try {
            if (this.treeView) {
                const parent = await this.getParent(uri.uri)
                // if (parent && !this.isExpandedNode.has(parent)) {
                if (parent) {
                    await this.didFocusString(SourcegraphUri.parse(parent), false, token)
                } else {
                    await this.getChildren(undefined)
                }
                if (token.isCancellationRequested) {
                    return
                }
                // log.appendLine(`didFocusString(${uri.uri})`)
                await this.treeView.reveal(uri.uri, {
                    focus: true,
                    select: isDestinationNode,
                    expand: !isDestinationNode,
                })
            }
        } catch (error) {
            log.error(`didFocusString(${uri.uri})`, error)
        }
    }

    private newTreeItem(
        uri: SourcegraphUri,
        parent: SourcegraphUri | undefined,
        parentChildrenCount: number
    ): vscode.TreeItem {
        const command = uri.isFile()
            ? {
                  command: 'extension.openFile',
                  title: 'Open file',
                  arguments: [uri.uri],
              }
            : undefined
        return {
            id: uri.uri,
            label: uri.treeItemLabel(parent),
            tooltip: uri.uri.replace('sourcegraph://', 'https://'),
            collapsibleState: uri.isFile()
                ? vscode.TreeItemCollapsibleState.None
                : parentChildrenCount === 1
                ? vscode.TreeItemCollapsibleState.Expanded
                : vscode.TreeItemCollapsibleState.Collapsed,
            command,
            resourceUri: vscode.Uri.parse(uri.uri),
            contextValue: uri.isFile() ? 'file' : 'directory',
        }
    }
}
