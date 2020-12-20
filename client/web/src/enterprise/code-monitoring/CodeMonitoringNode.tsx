import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useState, useCallback, useMemo } from 'react'
import * as H from 'history'
import { Observable, concat, of } from 'rxjs'
import { switchMap, catchError, startWith, takeUntil, tap, delay, mergeMap } from 'rxjs/operators'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { Link } from '../../../../shared/src/components/Link'
import { ErrorLike, isErrorLike, asError } from '../../../../shared/src/util/errors'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { CodeMonitorFields, ToggleCodeMonitorEnabledResult } from '../../graphql-operations'
import { CodeMonitoringProps } from '.'
import { sendTestEmail } from './backend'
import { AuthenticatedUser } from '../../auth'

export interface CodeMonitorNodeProps extends Pick<CodeMonitoringProps, 'toggleCodeMonitorEnabled'> {
    node: CodeMonitorFields
    location: H.Location
    authentictedUser: AuthenticatedUser
    showCodeMonitoringTestEmailButton: boolean
}

const LOADING = 'LOADING' as const

export const CodeMonitorNode: React.FunctionComponent<CodeMonitorNodeProps> = ({
    toggleCodeMonitorEnabled,
    location,
    node,
    authentictedUser,
    showCodeMonitoringTestEmailButton,
}: CodeMonitorNodeProps) => {
    const [enabled, setEnabled] = useState<boolean>(node.enabled)

    const [toggleMonitor, toggleMonitorOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent>) =>
                click.pipe(
                    switchMap(() => {
                        const toggleMonitor = toggleCodeMonitorEnabled(node.id, !enabled).pipe(
                            tap(
                                (
                                    idAndEnabled:
                                        | typeof LOADING
                                        | ErrorLike
                                        | ToggleCodeMonitorEnabledResult['toggleCodeMonitor']
                                ) => {
                                    if (idAndEnabled !== LOADING && !isErrorLike(idAndEnabled)) {
                                        setEnabled(idAndEnabled.enabled)
                                    }
                                }
                            ),
                            catchError(error => [asError(error)])
                        )
                        return concat(
                            of(LOADING).pipe(startWith(enabled), delay(300), takeUntil(toggleMonitor)),
                            toggleMonitor
                        )
                    })
                ),
            [node, enabled, setEnabled, toggleCodeMonitorEnabled]
        )
    )

    const [sendEmailRequest] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    mergeMap(() =>
                        sendTestEmail(node.trigger.id).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [node]
        )
    )

    const hasEnabledAction = useMemo(() => node.actions.nodes.filter(node => node.enabled).length > 0, [node.actions])

    return (
        <div className="card p-3 mb-2">
            <div className="d-flex justify-content-between align-items-center">
                <div className="d-flex flex-column">
                    <div className="font-weight-bold">{node.description}</div>
                    {/** TODO: Generate this text based on the type of action when new actions are added. */}
                    {node.actions.nodes.length > 0 && (
                        <div className="d-flex text-muted">
                            New search result → Sends email notifications{' '}
                            {showCodeMonitoringTestEmailButton &&
                                authentictedUser.siteAdmin &&
                                hasEnabledAction &&
                                node.enabled && (
                                    <button
                                        type="button"
                                        className="btn btn-link p-0 border-0 ml-2 test-send-test-email"
                                        onClick={sendEmailRequest}
                                    >
                                        Send test email
                                    </button>
                                )}
                        </div>
                    )}
                </div>
                <div className="d-flex flex-column">
                    <div className="d-flex">
                        {toggleMonitorOrError === LOADING && <LoadingSpinner className="icon-inline mr-2" />}
                        <div onClick={toggleMonitor} className="test-toggle-monitor-enabled">
                            <Toggle value={enabled} className="mr-3" disabled={toggleMonitorOrError === LOADING} />
                        </div>
                        <Link to={`${location.pathname}/${node.id}`}>Edit</Link>
                    </div>
                </div>
            </div>
            {isErrorLike(toggleMonitorOrError) && (
                <div className="alert alert-danger">Failed to toggle monitor: {toggleMonitorOrError.message}</div>
            )}
        </div>
    )
}
