import React, { useCallback, useMemo } from 'react'
import * as H from 'history'
import { Link } from '../../../../shared/src/components/Link'
import { Observable } from 'rxjs'
import { catchError, startWith, mergeMap, tap } from 'rxjs/operators'
import { asError } from '../../../../shared/src/util/errors'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { CodeMonitorFields } from '../../graphql-operations'

import { sendTestEmail } from './backend'
import { AuthenticatedUser } from '../../auth'

export interface CodeMonitorNodeProps {
    node: CodeMonitorFields
    location: H.Location
    authentictedUser: AuthenticatedUser
    showCodeMonitoringTestEmailButton: boolean
}

const LOADING = 'LOADING' as const

export const CodeMonitorNode: React.FunctionComponent<CodeMonitorNodeProps> = ({
    location,
    node,
    authentictedUser,
    showCodeMonitoringTestEmailButton,
}: CodeMonitorNodeProps) => {
    const [sendEmailRequest] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    tap(click => click.stopPropagation()),
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
        <Link to={`${location.pathname}/${node.id}`} className="code-monitoring-node card p-3">
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
                    <button className="btn btn-link" type="button">
                        Edit
                    </button>
                </div>
            </div>
        </Link>
    )
}
