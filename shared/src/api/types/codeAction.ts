import * as sourcegraph from 'sourcegraph'
import { CodeAction, Diagnostic, Range, Selection } from '@sourcegraph/extension-api-types'
import { fromDiagnostic, toDiagnostic } from './diagnostic'
import { WorkspaceEdit } from './workspaceEdit'
import { CodeActionsParams } from '../client/services/codeActions'
import { Range as RangeImpl, Selection as SelectionImpl } from '@sourcegraph/extension-api-classes'

export function fromCodeAction(codeAction: sourcegraph.CodeAction): CodeAction {
    return {
        ...codeAction,
        diagnostics: codeAction.diagnostics && codeAction.diagnostics.map(fromDiagnostic),
        edit: codeAction.edit && (codeAction.edit as WorkspaceEdit).toJSON(),
    }
}

export function toCodeAction(codeAction: CodeAction): sourcegraph.CodeAction {
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
