import { WorkspaceEdit } from 'vscode-languageserver-types'
import {
    ApplyWorkspaceEditParams,
    ApplyWorkspaceEditRequest,
    ApplyWorkspaceEditResponse,
    InitializeParams,
    ServerCapabilities,
} from '../../protocol'
import { Configuration, ConfigurationFeature } from '../configuration'
import { IConnection } from '../server'
import { WorkspaceFolders, WorkspaceFoldersFeature } from '../workspaceFolders'
import { Remote } from './common'

/**
 * Represents the workspace managed by the client.
 */
// tslint:disable-next-line:class-name
export interface _RemoteWorkspace extends Remote {
    /**
     * Applies a `WorkspaceEdit` to the workspace
     * @param param the workspace edit params.
     * @return a thenable that resolves to the `ApplyWorkspaceEditResponse`.
     */
    applyEdit(paramOrEdit: ApplyWorkspaceEditParams | WorkspaceEdit): Promise<ApplyWorkspaceEditResponse>
}

export type RemoteWorkspace = _RemoteWorkspace & Configuration & WorkspaceFolders

// tslint:disable-next-line:class-name
class _RemoteWorkspaceImpl implements _RemoteWorkspace {
    private _connection?: IConnection

    public attach(connection: IConnection): void {
        this._connection = connection
    }

    public get connection(): IConnection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public applyEdit(paramOrEdit: ApplyWorkspaceEditParams | WorkspaceEdit): Promise<ApplyWorkspaceEditResponse> {
        function isApplyWorkspaceEditParams(
            value: ApplyWorkspaceEditParams | WorkspaceEdit
        ): value is ApplyWorkspaceEditParams {
            return value && !!(value as ApplyWorkspaceEditParams).edit
        }

        const params: ApplyWorkspaceEditParams = isApplyWorkspaceEditParams(paramOrEdit)
            ? paramOrEdit
            : { edit: paramOrEdit }
        return this.connection.sendRequest(ApplyWorkspaceEditRequest.type, params)
    }
}

// tslint:disable-next-line:no-inferred-empty-object-type
export const RemoteWorkspaceImpl: new () => RemoteWorkspace = WorkspaceFoldersFeature(
    // tslint:disable-next-line:no-inferred-empty-object-type
    ConfigurationFeature(_RemoteWorkspaceImpl)
) as new () => RemoteWorkspace
