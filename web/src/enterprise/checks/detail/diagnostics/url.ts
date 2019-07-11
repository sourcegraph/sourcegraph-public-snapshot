import * as sourcegraph from 'sourcegraph'

export const urlToCheckDiagnosticGroup = (checkDiagnosticsURL: string, id: sourcegraph.DiagnosticGroup['id']): string =>
    `${checkDiagnosticsURL}/${id}`
