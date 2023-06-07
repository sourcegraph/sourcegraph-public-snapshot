import * as vscode from 'vscode'

import { getSingleLineRange } from '../services/InlineAssist'

import { CodyTaskState } from './utils'

const initState = new Map<string, vscode.CodeLens[]>()

export class FixupCodeLenses implements vscode.CodeLensProvider {
    private static provider: FixupCodeLenses
    private lenses = initState

    private _disposables: vscode.Disposable[] = []
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>()
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event
    /**
     * Create a code lens provider
     */
    constructor() {
        this.provideCodeLenses = this.provideCodeLenses.bind(this)
        this._disposables.push(vscode.languages.registerCodeLensProvider('*', this))
        // TODO (bea) update ranges on change
    }
    /**
     * Getter
     */
    public static get instance(): FixupCodeLenses {
        return (this.provider ??= new this())
    }
    /**
     * Create code lenses for each state
     */
    public provideCodeLenses(): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        // return all lenses from this.lenses map, expect state with fixed
        const lenses = this.lenses.values()
        return lenses ? Array.from(lenses).flat() : []
    }

    public set(id: string, state: CodyTaskState, range: vscode.Range): void {
        // All new lenses should be created as queued
        if (!this.lenses.has(id) && state !== CodyTaskState.queued) {
            return
        }
        if (state === CodyTaskState.fixed) {
            this.remove(id)
            return
        }
        const newRange = getSingleLineRange(range.start.line)
        const lens = getLensesByState(newRange, state, id)
        this.lenses.set(id, lens)
        this._onDidChangeCodeLenses.fire()
    }

    public remove(id: string): void {
        this.lenses.delete(id)
        this._onDidChangeCodeLenses.fire()
    }
    /**
     * Dispose the disposables
     */
    public dispose(): void {
        this.lenses = initState
        for (const disposable of this._disposables) {
            disposable.dispose()
        }
        this._disposables = []
    }
}

// Create Lenses based on state
function getLensesByState(codeLensRange: vscode.Range, state: CodyTaskState, id: string): vscode.CodeLens[] {
    switch (state) {
        case CodyTaskState.error: {
            const title = getErrorLens(codeLensRange)
            const diff = getDiffLens(codeLensRange, id)
            const discard = getDiscardLens(codeLensRange, id)
            return [title, diff, discard]
        }
        case CodyTaskState.applying: {
            const title = getApplyingLens(codeLensRange)
            return [title]
        }
        case CodyTaskState.pending: {
            const title = getAskingLens(codeLensRange)
            const cancel = getCancelLens(codeLensRange, id)
            return [title, cancel]
        }
        case CodyTaskState.done: {
            const title = getReadyLens(codeLensRange, id)
            const apply = getApplyLens(codeLensRange, id)
            const diff = getDiffLens(codeLensRange, id)
            const edit = getEditLens(codeLensRange, id)
            return [title, apply, edit, diff]
        }
        default:
            return []
    }
}

function getErrorLens(codeLensRange: vscode.Range): vscode.CodeLens {
    const lens = new vscode.CodeLens(codeLensRange)
    lens.command = {
        title: '⛔️ Fixup failed to apply',
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
