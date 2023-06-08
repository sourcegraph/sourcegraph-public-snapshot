import * as vscode from 'vscode'

import { getSingleLineRange } from '../services/InlineAssist'

import { CodyTaskState } from './utils'

// Create Lenses based on state
export function getLensesByState(id: string, state: CodyTaskState, range: vscode.Range): vscode.CodeLens[] {
    const codeLensRange = getSingleLineRange(range.start.line)
    switch (state) {
        case CodyTaskState.error: {
            const title = getErrorLens(codeLensRange)
            const diff = getDiffLens(codeLensRange, id)
            const discard = getDiscardLens(codeLensRange, id)
            return [title, diff, discard]
        }
        case CodyTaskState.applying: {
            const title = getApplyingLens(codeLensRange)
            const cancel = getCancelLens(codeLensRange, id)
            return [title, cancel]
        }
        case CodyTaskState.asking: {
            const title = getAskingLens(codeLensRange)
            const cancel = getCancelLens(codeLensRange, id)
            return [title, cancel]
        }
        case CodyTaskState.ready: {
            const title = getReadyLens(codeLensRange, id)
            const apply = getApplyLens(codeLensRange, id)
            const diff = getDiffLens(codeLensRange, id)
            const edit = getEditLens(codeLensRange, id)
            return [title, apply, edit, diff]
        }
        case CodyTaskState.marking: {
            const title = getMakingFixupsLens(codeLensRange)
            const follow = getFollowLens(codeLensRange, id)
            return [title, follow]
        }
        default:
            return []
    }
}

// List of lenses
// NOTE: code lens requires a command so we will use 'cody.focus' as a placeholder
// TODO (bea) send error messages to the chat UI so that they can see the task progress in the chat and chat history
function getErrorLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(warning) Fixup failed to apply',
        command: 'cody.focus',
    }
    return lens
}

function getAskingLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(sync~spin) Asking Cody...',
        command: 'cody.focus',
    }
    return lens
}

function getApplyingLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(sync~spin) Applying...',
        command: 'cody.focus',
    }
    return lens
}

function getCancelLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Cancel',
        command: 'cody.fixup.codelens.cancel',
        arguments: [id],
    }
    return lens
}

function getDiscardLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Discard',
        command: 'cody.fixup.codelens.discard',
        arguments: [id],
    }
    return lens
}

function getDiffLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Show Diff',
        command: 'cody.fixup.codelens.diff',
        arguments: [id],
    }
    return lens
}

function getReadyLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '✨ Fixup ready',
        command: 'cody.focus',
        arguments: [id],
    }
    return lens
}

function getEditLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Edit',
        command: 'cody.fixup.codelens.edit',
        arguments: [id],
    }
    return lens
}

function getApplyLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Apply',
        command: 'cody.fixup.codelens.apply',
        arguments: [id],
    }
    return lens
}

function getMakingFixupsLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(edit) Cody is making fixups...',
        command: 'cody.focus',
    }
    return lens
}
function getFollowLens(codeLensRange: vscode.Range, id: string): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: 'Follow',
        command: 'cody.fixup.open',
        arguments: [id],
    }
    return lens
}
