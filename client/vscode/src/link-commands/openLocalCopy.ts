import vscode from 'vscode'

/**
 * Open local copy of the remote file
 */

export async function openLocalCopy(): Promise<void> {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    // Get basePath from configuration
    const basePath = vscode.workspace.getConfiguration('sourcegraph').get<string>('basePath')
    const remotePath = editor.document.uri // => ex: {path: "/go/github.com/sourcegraph/sourcegraph@HEAD/-/blob/client/vscode/package.json"}
    const getPath = remotePath.path.split('/-/blob/')
    const repoNameWithBranch = getPath[0].split('/').pop() // ex: "/go/github.com/sourcegraph/sourcegraph@HEAD"
    const repoName = repoNameWithBranch !== undefined ? repoNameWithBranch.split('@').shift() : undefined
    const relativePath = getPath[1] // ex: "client/vscode/package.json"
    if (basePath !== undefined && repoName !== undefined) {
        const uri = vscode.Uri.file(`${basePath}${repoName}/${relativePath}`)
        const textDocument = await vscode.workspace.openTextDocument(uri)
        if (textDocument.fileName) {
            await vscode.window.showTextDocument(textDocument, {
                viewColumn: vscode.ViewColumn.Active,
            })
        }
    }
}
