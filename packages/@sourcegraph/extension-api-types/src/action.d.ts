import * as sourcegraph from 'sourcegraph'
import { Diagnostic } from './diagnostic'
import { WorkspaceEdit } from './workspaceEdit'

/**
 * An action.
 *
 * @see module:sourcegraph.Action
 */
export interface Action extends Pick<sourcegraph.Action, Exclude<keyof sourcegraph.Action, 'edit' | 'diagnostics'>> {
    readonly edit?: WorkspaceEdit
    readonly diagnostics?: Diagnostic[]
}
