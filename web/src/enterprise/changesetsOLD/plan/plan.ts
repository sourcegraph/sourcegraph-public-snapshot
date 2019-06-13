import * as sourcegraph from 'sourcegraph'

export interface ChangesetPlan {
    // TODO!(sqs): always assume only 1 operation
    operations: ChangesetPlanOperation[]
}

export interface ChangesetPlanOperation extends sourcegraph.Operation {}
