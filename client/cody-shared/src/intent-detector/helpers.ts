import vscode from 'vscode'
import execa from 'execa'
import path from 'path'

export function getProjectName(): string | undefined {
    const workspaceFolders = vscode.workspace.workspaceFolders
    if (!workspaceFolders) {
        return undefined
    }
    const projectFolders = workspaceFolders?.filter(folder => folder.uri.scheme === 'file')
    if (projectFolders.length === 1) {
        return projectFolders[0].name
    }
    const activeTextEditor = vscode.window.activeTextEditor
    if (!activeTextEditor) {
        return undefined
    }
    const activeFileUri = activeTextEditor.document.uri
    for (const projectFolder of projectFolders) {
        if (activeFileUri.fsPath.startsWith(projectFolder.uri.fsPath)) {
            return projectFolder.name
        }
    }
    return undefined
}

export async function getRepoName(): Promise<string | undefined> {
    const repoRoot = await getRepoRoot()
    const { stdout } = await execa('git', ['remote', 'get-url', 'origin'], { cwd: repoRoot }) // TODO: Make "origin" dynamic
    return stdout
}

async function getRepoRoot(): Promise<string | undefined> {
    const filePath = getCurrentFilePath()
    if (!filePath) {
        return undefined
    }

    // Determine repository root directory.
    const fileDirectory = path.dirname(filePath)
    const { stdout: repoRoot } = await execa('git', ['rev-parse', '--show-toplevel'], { cwd: fileDirectory })
    return repoRoot
}

export async function getCurrentBranchName(): Promise<string | undefined> {
    const repoRoot = await getRepoRoot()

    const { stdout } = await execa('git', ['rev-parse', '--abbrev-ref', 'HEAD'], { cwd: repoRoot })
    return stdout !== 'HEAD' ? stdout : undefined
}

export function getCurrentFilePath(): string | undefined {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        return undefined
    }
    return editor.document.uri.fsPath
}

export function getOpenFilePaths(): string[] {
    const tabs = vscode.window.tabGroups.activeTabGroup.tabs
    return tabs.map(tab => tab.input ? (tab.input as {
        uri: string | undefined
    }).uri : undefined).filter(Boolean) as string[]
}
