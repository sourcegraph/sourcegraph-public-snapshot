import { useCallback, useState } from 'react'
import { DiagnosticQuery } from 'sourcegraph'
import { DiagnosticWithType } from '../../../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../../../shared/src/api/types/action'
import { isDiagnosticQueryEqual } from '../../../../../../shared/src/api/types/diagnostic'
import { ChangesetPlan, ChangesetPlanOperation } from '../../../changesets/plan/plan'
import { diagnosticID, diagnosticQueryForSingleDiagnostic } from '../../../threads/detail/backend'

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
            const diagnosticQuery = diagnosticQueryForSingleDiagnostic(diagnostic)
            const existingIndex = changesetPlan.operations.findIndex(
                op => op.diagnostics && isDiagnosticQueryEqual(op.diagnostics, diagnosticQuery)
            )

            if (action) {
                if (!action.computeEdit) {
                    throw new Error('TODO!(sqs) shouldnt be undefined')
                }
                const op: ChangesetPlanOperation = {
                    message: action.title,
                    diagnostics: diagnosticQuery,
                    editCommand: action.computeEdit,
                }

                // Add or update this entry.
                if (existingIndex >= 0) {
                    changesetPlan.operations[existingIndex] = op
                } else {
                    // Insert operations on single diagnostics at the start so that they are applied
                    // even if there are broader operations at the end that apply to more than 1
                    // diagnostic.
                    changesetPlan.operations.unshift(op)
                }
            } else {
                // Remove this entry.
                if (existingIndex >= 0) {
                    changesetPlan.operations.splice(existingIndex, 1)
                } else {
                    console.error('TODO!(sqs): handle this error')
                }
            }
            setChangesetPlan({ ...changesetPlan, operations: [...changesetPlan.operations] })
        },
        [changesetPlan]
    )

    return { changesetPlan, onChangesetPlanDiagnosticActionSet }
}
