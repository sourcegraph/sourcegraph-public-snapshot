import * as sourcegraph from 'sourcegraph'
import { Diagnostic } from './diagnostic'

/**
 * An action.
 *
 * @see module:sourcegraph.Action
 */
export interface Action extends Pick<sourcegraph.Action, Exclude<keyof sourcegraph.Action, 'edit' | 'diagnostics'>> {
    readonly edit?: any // TODO!(sqs): use WorkspaceEdit type
    readonly diagnostics?: Diagnostic[]
}

// TODO!(sqs): move Action out of the types package and into shared/src/api/types because we dont want to move all of WorkspaceEdit into this external package
