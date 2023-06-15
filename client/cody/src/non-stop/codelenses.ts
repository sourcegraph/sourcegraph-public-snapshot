import * as vscode from 'vscode'

import { getSingleLineRange } from '../services/InlineAssist'

import { FixupTask } from './FixupTask'
import { CodyTaskState } from './utils'

// Create Lenses based on state
export function getLensesForTask(task: FixupTask): vscode.CodeLens[] {
    const codeLensRange = getSingleLineRange(task.selectionRange.start.line)
    switch (task.state) {
        case CodyTaskState.waiting: {
            const title = getWaitingLens(codeLensRange)
            const cancel = getCancelLens(codeLensRange, task.id)
            return [title, cancel]
        }
        case CodyTaskState.asking: {
            const title = getAskingLens(codeLensRange)
            const cancel = getCancelLens(codeLensRange, task.id)
            return [title, cancel]
        }
        case CodyTaskState.ready: {
            const title = getReadyLens(codeLensRange, task.id)
            const apply = getApplyLens(codeLensRange, task.id)
            const diff = getDiffLens(codeLensRange, task.id)
            return [title, apply, diff]
        }
        case CodyTaskState.applying: {
            const title = getApplyingLens(codeLensRange)
            return [title]
        }
        case CodyTaskState.error: {
            const title = getErrorLens(codeLensRange)
            const discard = getDiscardLens(codeLensRange, task.id)
            return [title, discard]
        }
        default:
            return []
    }
}

// List of lenses
// TODO: Replace cody.focus with appropriate tasks
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

function getWaitingLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '$(sync) Asking Cody...',
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
        command: 'cody.fixup.codelens.cancel',
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
        title: '$(pencil) Fixup ready',
        command: 'cody.fixup.codelens.apply',
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
