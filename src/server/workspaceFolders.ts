import { Unsubscribable } from 'rxjs'
import { Emitter, Event } from '../jsonrpc2/events'
import {
    ClientCapabilities,
    DidChangeWorkspaceFoldersNotification,
    WorkspaceFolder,
    WorkspaceFoldersChangeEvent,
    WorkspaceFoldersRequest,
} from '../protocol'
import { _RemoteWorkspace } from './features/workspace'
import { Feature } from './server'

export interface WorkspaceFolders {
    getWorkspaceFolders(): Promise<WorkspaceFolder[] | null>
    onDidChangeWorkspaceFolders: Event<WorkspaceFoldersChangeEvent>
}

export const WorkspaceFoldersFeature: Feature<_RemoteWorkspace, WorkspaceFolders> = Base =>
    class extends Base {
        private _onDidChangeWorkspaceFolders?: Emitter<WorkspaceFoldersChangeEvent>
        private _unregistration?: Promise<Unsubscribable>

        public initialize(capabilities: ClientCapabilities): void {
            const workspaceCapabilities = capabilities.workspace
            if (workspaceCapabilities && workspaceCapabilities.workspaceFolders) {
                this._onDidChangeWorkspaceFolders = new Emitter<WorkspaceFoldersChangeEvent>()
                this.connection.onNotification(DidChangeWorkspaceFoldersNotification.type, params => {
                    this._onDidChangeWorkspaceFolders!.fire(params.event)
                })
            }
        }

        public getWorkspaceFolders(): Promise<WorkspaceFolder[] | null> {
            return this.connection.sendRequest(WorkspaceFoldersRequest.type, null)
        }

        public get onDidChangeWorkspaceFolders(): Event<WorkspaceFoldersChangeEvent> {
            if (!this._onDidChangeWorkspaceFolders) {
                throw new Error("Client doesn't support sending workspace folder change events.")
            }
            if (!this._unregistration) {
                this._unregistration = this.connection.client.register(DidChangeWorkspaceFoldersNotification.type)
            }
            return this._onDidChangeWorkspaceFolders.event
        }
    }
