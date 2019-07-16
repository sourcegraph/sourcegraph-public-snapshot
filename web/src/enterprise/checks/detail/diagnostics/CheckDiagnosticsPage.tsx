import H from 'history'
import React, { useCallback } from 'react'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticWithType } from '../../../../../../shared/src/api/client/services/diagnosticService'
import { Action } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { propertyIsDefined } from '../../../../../../shared/src/util/types'
import { ThemeProps } from '../../../../theme'
import { DiagnosticInfo, diagnosticQueryMatcher } from '../../../threads/detail/backend'
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
    const { changesetPlan, onChangesetPlanDiagnosticActionSet, onChangesetPlanBatchActionClick } = useChangesetPlan()

    const onActionSelect = useCallback(
        (diagnostic: DiagnosticInfo, action: Action | null) => {
            onChangesetPlanDiagnosticActionSet(diagnostic, action)
        },
        [onChangesetPlanDiagnosticActionSet]
    )

    const getSelectedActionForDiagnostic = useCallback(
        (diagnostic: DiagnosticWithType) =>
            changesetPlan.operations
                .filter(propertyIsDefined('diagnostics'))
                .find(op => diagnosticQueryMatcher(op.diagnostics)(diagnostic)) || null,
        [changesetPlan.operations]
    )

    return (
        <div className={`check-diagnostics-page flex-1 d-flex flex-column ${className}`}>
            <DiagnosticsListPage
                {...props}
                getSelectedActionForDiagnostic={getSelectedActionForDiagnostic}
                onActionSelect={onActionSelect}
                changesetPlan={changesetPlan}
                onChangesetPlanBatchActionClick={onChangesetPlanBatchActionClick}
                checkProvider={checkProvider}
                className="flex-1"
                defaultQuery={`type:${checkID.type}`}
            />
            <div className="check-diagnostics-page__bar border-top">
                <DiagnosticsChangesetsBar {...props} changesetPlan={changesetPlan} />
            </div>
        </div>
    )
}
