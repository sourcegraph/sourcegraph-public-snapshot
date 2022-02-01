import * as H from 'history'
import React, { useCallback, useEffect, useMemo } from 'react'
import { Observable } from 'rxjs'

import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { convertActionsForCreate } from './action-converters'
import { createCodeMonitor as _createCodeMonitor } from './backend'
import { CodeMonitorForm } from './components/CodeMonitorForm'

interface CreateCodeMonitorPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser

    createCodeMonitor?: typeof _createCodeMonitor
}

const AuthenticatedCreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = ({
    authenticatedUser,
    history,
    location,
    createCodeMonitor = _createCodeMonitor,
}) => {
    const triggerQuery = useMemo(() => new URLSearchParams(location.search).get('trigger-query') ?? undefined, [
        location.search,
    ])

    const description = useMemo(() => new URLSearchParams(location.search).get('description') ?? undefined, [
        location.search,
    ])

    useEffect(
        () =>
            eventLogger.logViewEvent('CreateCodeMonitorPage', {
                hasTriggerQuery: !!triggerQuery,
                hasDescription: !!description,
            }),
        [triggerQuery, description]
    )

    const createMonitorRequest = useCallback(
        (codeMonitor: CodeMonitorFields): Observable<Partial<CodeMonitorFields>> => {
            eventLogger.log('CreateCodeMonitorFormSubmitted')
            return createCodeMonitor({
                monitor: {
                    namespace: authenticatedUser.id,
                    description: codeMonitor.description,
                    enabled: codeMonitor.enabled,
                },
                trigger: { query: codeMonitor.trigger.query },

                actions: convertActionsForCreate(codeMonitor.actions.nodes, authenticatedUser.id),
            })
        },
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
                description={description}
                submitButtonLabel="Create code monitor"
            />
        </div>
    )
}

export const CreateCodeMonitorPage = withAuthenticatedUser(AuthenticatedCreateCodeMonitorPage)
