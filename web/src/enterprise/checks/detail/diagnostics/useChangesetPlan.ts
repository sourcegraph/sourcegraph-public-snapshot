import { Action } from '@sourcegraph/extension-api-types'
import { useCallback, useState } from 'react'
import { DiagnosticWithType } from '../../../../../../shared/src/api/client/services/diagnosticService'
import { ChangesetPlan, ChangesetPlanDiagnosticAction, ChangesetPlanOperation } from '../../../changesets/plan/plan'
import { diagnosticID } from '../../../threads/detail/backend'

export interface ChangesetPlanProps {
    changesetPlan: ChangesetPlan
    // onChangesetPlanChange: (plan: ChangesetPlan) => void
    onChangesetPlanDiagnosticActionSet: (diagnostic: DiagnosticWithType, action: Action | null) => void
}

/**
 * A React hook that manages state for a new {@link ChangesetPlan} that is being created.
 */
export const useChangesetPlan = (): ChangesetPlanProps => {
    const [changesetPlan, setChangesetPlan] = useState<ChangesetPlan>({
        operations: [],
    })

    const onChangesetPlanDiagnosticActionSet = useCallback<ChangesetPlanProps['onChangesetPlanDiagnosticActionSet']>(
        (diagnostic, action) => {
            let op: ChangesetPlanOperation
            if (changesetPlan.operations.length === 0) {
                op = { diagnosticQuery: 'TODO!(sqs)', diagnosticActions: [] }
                changesetPlan.operations.push(op)
            } else {
                // TODO!(sqs): always assume adding to or removing from 1st op, and there is never more than 1 op
                op = changesetPlan.operations[0]
            }

            const diagID = diagnosticID(diagnostic)
            const existingIndex = op.diagnosticActions.findIndex(d => d.diagnosticID === diagID)
            if (action) {
                // Add or update this entry.
                const newEntry: ChangesetPlanDiagnosticAction = { diagnosticID: diagID, action }
                if (existingIndex >= 0) {
                    op.diagnosticActions[existingIndex] = newEntry
                } else {
                    op.diagnosticActions.push(newEntry)
                }
            } else {
                // Remove this entry.
                if (existingIndex >= 0) {
                    op.diagnosticActions.splice(existingIndex, 1)
                } else {
                    console.error('TODO!(sqs): handle this error')
                }
            }
            setChangesetPlan({ ...changesetPlan })
        },
        [changesetPlan]
    )

    return { changesetPlan, onChangesetPlanDiagnosticActionSet }
}
