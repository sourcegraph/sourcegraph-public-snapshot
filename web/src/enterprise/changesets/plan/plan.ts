import * as sourcegraph from 'sourcegraph'
import { Action } from '../../../../../shared/src/api/types/action'

export interface ChangesetPlan {
    // TODO!(sqs): always assume only 1 operation
    operations: ChangesetPlanOperation[]
}

export interface ChangesetPlanOperation {
    diagnosticQuery: sourcegraph.DiagnosticQuery | 'TODO!(sqs)'
    diagnosticActions: ChangesetPlanDiagnosticAction[]
}

export interface ChangesetPlanDiagnosticAction {
    // TODO!(sqs): think about how to do diagnostic identifiers
    diagnosticID: string
    action: Action
}
