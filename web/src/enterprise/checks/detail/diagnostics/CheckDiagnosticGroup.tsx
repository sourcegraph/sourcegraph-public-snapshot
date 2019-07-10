import { DiagnosticSeverity, NotificationType } from '@sourcegraph/extension-api-classes'
import { sortBy } from 'lodash'
import LightbulbIcon from 'mdi-react/LightbulbIcon'
import React, { useCallback, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { ActionType } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { useOnActionClickCallback } from '../../../actions/useOnActionClickCallback'
import { ChangesetCreationStatus, createChangeset } from '../../../changesets/preview/backend'
import { NotificationActions } from '../../../notifications/actions/NotificationActions'
import { NotificationTypeIcon } from '../../../notifications/NotificationTypeIcon'
import {
    ChangesetButtonOrLinkExistingChangeset,
    PENDING_CREATION,
} from '../../../tasks/list/item/ChangesetButtonOrLink'
import { useDiagnostics } from './useDiagnostics'

interface Props extends ExtensionsControllerProps {
    diagnosticGroup: sourcegraph.DiagnosticGroup
    className?: string
    contentClassName?: string
}

const LOADING = 'loading' as const

/**
 * A diagnostic group associated with a status.
 */
export const CheckDiagnosticGroup: React.FunctionComponent<Props> = ({
    diagnosticGroup,
    className = '',
    contentClassName = '',
    extensionsController,
}) => {
    const diagnosticsOrError = useDiagnostics(extensionsController, diagnosticGroup.query)

    const diagnosticSeverityCountMap = new Map<sourcegraph.DiagnosticSeverity, number>()
    if (diagnosticsOrError !== LOADING && !isErrorLike(diagnosticsOrError)) {
        for (const diag of diagnosticsOrError) {
            diagnosticSeverityCountMap.set(diag.severity, (diagnosticSeverityCountMap.get(diag.severity) || 0) + 1)
        }
    }
    const diagnosticSeverityCounts = sortBy([...diagnosticSeverityCountMap.entries()], 0)

    const [createdChangesetOrLoading, setCreatedChangesetOrLoading] = useState<ChangesetButtonOrLinkExistingChangeset>(
        null
    )
    const onPlanActionClick = useCallback(
        async (plan: ActionType['plan'], creationStatus: ChangesetCreationStatus) => {
            setCreatedChangesetOrLoading(PENDING_CREATION)
            try {
                setCreatedChangesetOrLoading(
                    await createChangeset(
                        { extensionsController },
                        {
                            title: plan.plan.operations[0].command.title,
                            contents: diagnosticGroup.message,
                            status: creationStatus,
                            plan: plan.plan,
                            changesetActionDescriptions: [
                                {
                                    title: plan.plan.title,
                                    timestamp: Date.now(),
                                    user: 'sqs',
                                },
                            ],
                        }
                    )
                )
            } catch (err) {
                setCreatedChangesetOrLoading(null)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController, diagnosticGroup.message]
    )
    const isPlanLoading = createdChangesetOrLoading === LOADING

    const [onCommandActionClick, isCommandLoading] = useOnActionClickCallback(extensionsController)

    const disabled = isPlanLoading || isCommandLoading

    return (
        <div className={`status-notification ${className}`}>
            <div className={`d-flex align-items-start ${contentClassName}`}>
                <div className="flex-1">
                    <div className="d-flex align-items-start justify-content-between">
                        <h3 className="mb-0 font-weight-normal">{diagnosticGroup.name}</h3>
                        {diagnosticSeverityCounts.map(
                            ([diagnosticSeverity, count]) =>
                                count > 0 && (
                                    <span key={diagnosticSeverity}>
                                        <DiagnosticSeverityIcon
                                            severity={diagnosticSeverity}
                                            className="icon-inline h3 mb-0 mr-3 flex-0"
                                        />
                                        {count}
                                    </span>
                                )
                        )}
                    </div>
                    {/* TODO!(sqs) diagnosticGroup.actions && false && (
                        <NotificationActions
                            actions={diagnosticGroup.actions}
                            onPlanActionClick={onPlanActionClick}
                            onCommandActionClick={onCommandActionClick}
                            existingChangeset={createdChangesetOrLoading}
                            disabled={disabled}
                            className="mt-4"
                        />
                    )*/}
                </div>
            </div>
        </div>
    )
}
