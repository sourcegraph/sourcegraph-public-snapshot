import * as H from 'history'
import React, { useCallback, useEffect, useMemo } from 'react'
import { Observable } from 'rxjs'

import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorFields, MonitorEmailPriority } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { createCodeMonitor as _createCodeMonitor } from './backend'
import { CodeMonitorForm } from './components/CodeMonitorForm'

export interface CreateCodeMonitorPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser

    createCodeMonitor?: typeof _createCodeMonitor
}

export const CreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = ({
    authenticatedUser,
    history,
    location,
    createCodeMonitor = _createCodeMonitor,
}) => {
    const triggerQuery = useMemo(() => new URLSearchParams(location.search).get('trigger-query') ?? undefined, [
        location.search,
    ])
    useEffect(() => eventLogger.logViewEvent('CreateCodeMonitorPage', { hasTriggerQuery: !!triggerQuery }), [
        triggerQuery,
    ])

    const createMonitorRequest = useCallback(
        (codeMonitor: CodeMonitorFields): Observable<Partial<CodeMonitorFields>> =>
            createCodeMonitor({
                monitor: {
                    namespace: authenticatedUser.id,
                    description: codeMonitor.description,
                    enabled: codeMonitor.enabled,
                },
                trigger: { query: codeMonitor.trigger.query },

                actions: codeMonitor.actions.nodes.map(action => ({
                    email: {
                        enabled: action.enabled,
                        priority: MonitorEmailPriority.NORMAL,
                        recipients: [authenticatedUser.id],
                        header: '',
                    },
                })),
            }),
        [authenticatedUser.id, createCodeMonitor]
    )

    return (
        <div className="container col-8">
            <PageTitle title="Create new code monitor" />
            <PageHeader
                path={[{ icon: CodeMonitoringLogo, to: '/code-monitoring' }, { text: 'Create code monitor' }]}
                description={
                    <>
                        Code monitors watch your code for specific triggers and run actions in response.{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more
                        </a>
                    </>
                }
            />
            <CodeMonitorForm
                history={history}
                location={location}
                authenticatedUser={authenticatedUser}
                onSubmit={createMonitorRequest}
                triggerQuery={triggerQuery}
                submitButtonLabel="Create code monitor"
            />
        </div>
    )
}
