import * as vscode from 'vscode'

import { getLensesByState } from './codelenses'
import { taskID } from './FixupTask'
import { CodyTaskState } from './utils'

const initState = new Map<taskID, vscode.CodeLens[]>()

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
    }
    /**
     * Getter
     * TODO (bea) Refactor this to make it not global
     */
    public static get instance(): FixupCodeLenses {
        return (this.provider ??= new this())
    }
    /**
     * This method is called by vscode to get the code lenses on every refresh
     * To update a lens, use the set method to replace the exisiting one with a new lens by the same id
     */
    public provideCodeLenses(): vscode.CodeLens[] | Thenable<vscode.CodeLens[]> {
        // return all lenses from this.lenses map, expect state with fixed
        const lenses = this.lenses.values()
        return lenses ? Array.from(lenses).flat() : []
    }

    public set(id: string, state: CodyTaskState, range: vscode.Range): void {
        if (state === CodyTaskState.fixed) {
            this.remove(id)
            return
        }
        const lens = getLensesByState(id, state, range)
        this.lenses.set(id, lens)
        this.refresh()
    }

    public remove(id: string): void {
        this.lenses.delete(id)
        this.refresh()
    }

    public refresh(): void {
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
