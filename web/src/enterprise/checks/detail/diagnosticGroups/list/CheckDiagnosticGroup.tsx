import { DiagnosticSeverity, NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import { sortBy } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import LightbulbIcon from 'mdi-react/LightbulbIcon'
import MenuRightIcon from 'mdi-react/MenuRightIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import React, { useCallback, useState } from 'react'
import { Link } from 'react-router-dom'
import * as sourcegraph from 'sourcegraph'
import { ActionType } from '../../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../../diagnostics/components/DiagnosticSeverityIcon'
import { ThemeProps } from '../../../../../theme'
import { useOnActionClickCallback } from '../../../../actions/useOnActionClickCallback'
import { ChangesetCreationStatus, createChangeset } from '../../../../changesets/preview/backend'
import { NotificationActions } from '../../../../notifications/actions/NotificationActions'
import { NotificationTypeIcon } from '../../../../notifications/NotificationTypeIcon'
import { DiagnosticsList } from '../../../../tasks/list/DiagnosticsList'
import {
    ChangesetButtonOrLinkExistingChangeset,
    PENDING_CREATION,
} from '../../../../tasks/list/item/ChangesetButtonOrLink'
import { useDiagnostics } from '../detail/useDiagnostics'
import { urlToCheckDiagnosticGroup } from '../url'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    diagnosticGroup: sourcegraph.DiagnosticGroup
    isExpanded: boolean
    checkDiagnosticsURL: string

    className?: string
    contentClassName?: string
    history: H.History
    location: H.Location
}

const LOADING = 'loading' as const

/**
 * A diagnostic group associated with a status.
 */
export const CheckDiagnosticGroup: React.FunctionComponent<Props> = ({
    diagnosticGroup,
    isExpanded,
    checkDiagnosticsURL,
    className = '',
    contentClassName = '',
    extensionsController,
    ...props
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

    const url = urlToCheckDiagnosticGroup(checkDiagnosticsURL, diagnosticGroup.id)
    const toggleIsExpandedUrl = isExpanded ? checkDiagnosticsURL : url
    const ToggleIsExpandedIcon = isExpanded ? MenuDownIcon : MenuRightIcon

    return (
        <div className={`check-diagnostic-group ${className}`}>
            <div className={contentClassName}>
                <header>
                    <div className="d-flex align-items-center position-relative">
                        <Link to={toggleIsExpandedUrl} className="btn btn-link px-0 stretched-link">
                            <ToggleIsExpandedIcon className="icon-inline" />
                        </Link>
                        {diagnosticSeverityCounts.map(
                            ([diagnosticSeverity, count]) =>
                                count > 0 && (
                                    <span key={diagnosticSeverity} className="d-flex align-items-center mr-3 h3 mb-0">
                                        <DiagnosticSeverityIcon
                                            severity={diagnosticSeverity}
                                            className="icon-inline mb-0 mr-1 small"
                                        />
                                        {count}
                                    </span>
                                )
                        )}
                        <h3 className="mb-0 font-weight-normal">
                            <Link to={url}>{diagnosticGroup.name}</Link>
                        </h3>
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
                </header>
                <div className="d-flex align-items-center mt-3">actions</div>
            </div>
            {isExpanded && (
                <DiagnosticsList
                    {...props}
                    diagnosticsOrError={diagnosticsOrError}
                    className="mb-5"
                    itemClassName="container-fluid"
                    extensionsController={extensionsController}
                />
            )}
        </div>
    )
}
