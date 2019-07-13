import H from 'history'
import React, { useCallback } from 'react'
import * as sourcegraph from 'sourcegraph'
import { Action } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../theme'
import { DiagnosticInfo } from '../../../threads/detail/backend'
import { CheckAreaContext } from '../CheckArea'
import { DiagnosticsChangesetsBar } from './changesets/DiagnosticsChangesetsBar'
import { DiagnosticsListPage } from './DiagnosticsListPage'
import { useChangesetPlan } from './useChangesetPlan'

interface Props
    extends Pick<CheckAreaContext, 'checkID' | 'checkProvider' | 'checkInfo'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The check diagnostics page.
 */
export const CheckDiagnosticsPage: React.FunctionComponent<Props> = ({
    checkID,
    checkProvider,
    checkInfo,
    className = '',
    ...props
}) => {
    const { changesetPlan, onChangesetPlanDiagnosticActionSet } = useChangesetPlan()
    const baseDiagnosticQuery: sourcegraph.DiagnosticQuery = { type: checkID.type }

    const selectedActions: { [diagnosticID: string]: Action | undefined } = {}
    for (const op of changesetPlan.operations) {
        // TODO!(sqs): assumes only 1 op
        for (const entry of op.diagnosticActions) {
            selectedActions[entry.diagnosticID] = entry.action
        }
    }
    const onActionSelect = useCallback(
        (diagnostic: DiagnosticInfo, action: Action | null) => {
            onChangesetPlanDiagnosticActionSet(diagnostic, action)
        },
        [onChangesetPlanDiagnosticActionSet]
    )

    return (
        <div className={`check-diagnostics-page ${className}`}>
            <DiagnosticsListPage
                {...props}
                baseDiagnosticQuery={baseDiagnosticQuery}
                selectedActions={selectedActions}
                onActionSelect={onActionSelect}
                checkProvider={checkProvider}
            />
            <div className="check-diagnostics-page__bar border-top">
                <DiagnosticsChangesetsBar
                    {...props}
                    changesetPlan={changesetPlan}
                    onChangesetPlanDiagnosticActionSet={onChangesetPlanDiagnosticActionSet}
                    className=""
                />
            </div>
        </div>
    )
}
