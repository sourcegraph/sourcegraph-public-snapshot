import * as sourcegraph from 'sourcegraph'
import { Diagnostic } from './diagnostic'
import { WorkspaceEdit } from './workspaceEdit'

/**
 * A code action.
 *
 * @see module:sourcegraph.CodeAction
 */
export interface CodeAction
    extends Pick<sourcegraph.CodeAction, Exclude<keyof sourcegraph.CodeAction, 'edit' | 'diagnostics'>> {
    readonly edit?: WorkspaceEdit
    readonly diagnostics?: Diagnostic[]
}
