import * as vscode from 'vscode'

export enum CodyTaskState {
    'idle' = 0,
    'queued' = 1,
    'pending' = 2,
    'stopped' = 3,
    'done' = 4,
    'error' = 5,
}

export type CodyTaskIcon = {
    [key in CodyTaskState]: {
        id: string
        icon: string
    }
}
/**
 * Icon for each task state
 */
export const fixupTaskIcon: CodyTaskIcon = {
    [CodyTaskState.idle]: {
        id: 'idle',
        icon: 'smiley',
    },
    [CodyTaskState.pending]: {
        id: 'pending',
        icon: 'sync~spin',
    },
    [CodyTaskState.done]: {
        id: 'done',
        icon: 'issue-closed',
    },
    [CodyTaskState.error]: {
        id: 'error',
        icon: 'stop',
    },
    [CodyTaskState.queued]: {
        id: 'queue',
        icon: 'clock',
    },
    [CodyTaskState.stopped]: {
        id: 'removed',
        icon: 'circle-slash',
    },
}
/**
 * For tracking lines diff
 */
export async function editDocByUri(
    uri: vscode.Uri,
    lines: { start: number; end: number },
    content: string
): Promise<vscode.Range> {
    // Highlight from the start line to the length of the replacement content
    const lineDiff = content.split('\n').length - 2
    const document = await vscode.workspace.openTextDocument(uri)
    const edit = new vscode.WorkspaceEdit()
    const range = new vscode.Range(lines.start, 0, lines.end + 1, 0)
    edit.delete(document.uri, range)
    edit.insert(document.uri, new vscode.Position(lines.start, 0), content)
    await vscode.workspace.applyEdit(edit)
    return new vscode.Range(lines.start, 0, lines.start + lineDiff, 0)
}
/**
 * Get the last part of the file path after the last slash
 */
export function getFileNameAfterLastDash(filePath: string): string {
    const lastDashIndex = filePath.lastIndexOf('/')
    if (lastDashIndex === -1) {
        return filePath
    }
    return filePath.slice(lastDashIndex + 1)
}
