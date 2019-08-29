import * as sourcegraph from 'sourcegraph'
import { Diagnostic, Range, Selection } from '@sourcegraph/extension-api-types'
import { fromDiagnostic, toDiagnostic } from './diagnostic'
import { WorkspaceEdit, SerializedWorkspaceEdit } from './workspaceEdit'
import { CodeActionsParams } from '../client/services/codeActions'
import { Range as RangeImpl, Selection as SelectionImpl } from '@sourcegraph/extension-api-classes'

/**
 * An action.
 *
 * @see module:sourcegraph.Action
 */
export interface Action extends Pick<sourcegraph.Action, Exclude<keyof sourcegraph.Action, 'edit' | 'diagnostics'>> {
    readonly edit?: SerializedWorkspaceEdit // TODO!(sqs): use WorkspaceEdit type
    readonly diagnostics?: Diagnostic[]
}

export function fromAction(codeAction: sourcegraph.Action): Action {
    return {
        ...codeAction,
        diagnostics: codeAction.diagnostics && codeAction.diagnostics.map(fromDiagnostic),
        edit: codeAction.edit && (codeAction.edit as WorkspaceEdit).toJSON(),
    }
}

export function toAction(codeAction: Action): sourcegraph.Action {
    return {
        ...codeAction,
        diagnostics: codeAction.diagnostics && codeAction.diagnostics.map(toDiagnostic),
        edit: codeAction.edit && WorkspaceEdit.fromJSON(codeAction.edit),
    }
}

export interface PlainCodeActionsParams extends Pick<CodeActionsParams, 'textDocument'> {
    range: Range | Selection
    context: { diagnostics: Diagnostic[] }
}

export function fromCodeActionsParams(params: CodeActionsParams): PlainCodeActionsParams {
    return {
        ...params,
        range: (params.range as any).toJSON(),
        context: { ...params.context, diagnostics: params.context.diagnostics.map(fromDiagnostic) },
    }
}

export function toCodeActionsParams(params: PlainCodeActionsParams): CodeActionsParams {
    return {
        ...params,
        range: RangeImpl.isRange(params.range)
            ? RangeImpl.fromPlain(params.range)
            : SelectionImpl.fromPlain(params.range),
        context: { ...params.context, diagnostics: params.context.diagnostics.map(toDiagnostic) },
    }
}

export function isCommandOnlyAction(
    action: Action | sourcegraph.Action
): action is Required<Pick<Action | sourcegraph.Action, 'title' | 'command'>> {
    return !action.edit && !action.computeEdit && (!action.diagnostics || action.diagnostics.length === 0)
}
