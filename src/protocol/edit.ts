import { WorkspaceEdit } from 'vscode-languageserver-types'
import { RequestType } from '../jsonrpc2/messages'

/**
 * The parameters passed via a apply workspace edit request.
 */
export interface ApplyWorkspaceEditParams {
    /**
     * An optional label of the workspace edit. This label is
     * presented in the user interface for example on an undo
     * stack to undo the workspace edit.
     */
    label?: string

    /**
     * The edits to apply.
     */
    edit: WorkspaceEdit
}

/**
 * A response returned from the apply workspace edit request.
 */
export interface ApplyWorkspaceEditResponse {
    /**
     * Indicates whether the edit was applied or not.
     */
    applied: boolean
}

/**
 * A request sent from the server to the client to modified certain resources.
 */
export namespace ApplyWorkspaceEditRequest {
    export const type = new RequestType<ApplyWorkspaceEditParams, ApplyWorkspaceEditResponse, void, void>(
        'workspace/applyEdit'
    )
}
