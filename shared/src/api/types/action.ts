import * as sourcegraph from 'sourcegraph'
import { CodeAction } from '@sourcegraph/extension-api-types'
import { fromDiagnostic, toDiagnostic } from './diagnostic'
import { WorkspaceEdit } from './workspaceEdit'

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
