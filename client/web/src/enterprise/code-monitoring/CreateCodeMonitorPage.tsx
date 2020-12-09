import * as H from 'history'
import React, { useCallback, useMemo } from 'react'
import { Observable } from 'rxjs'
import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorFields, MonitorEmailPriority } from '../../graphql-operations'
import { createCodeMonitor } from './backend'
import { CodeMonitorForm } from './components/CodeMonitorForm'

export interface CreateCodeMonitorPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser
}

export const CreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = props => {
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Create Code Monitor',
                element: <>Create new code monitor</>,
            }),
            []
        )
    )

    const createMonitorRequest = useCallback(
        (codeMonitor: CodeMonitorFields): Observable<Partial<CodeMonitorFields>> =>
            createCodeMonitor({
                monitor: {
                    namespace: props.authenticatedUser.id,
                    description: codeMonitor.description,
                    enabled: codeMonitor.enabled,
                },
                trigger: { query: codeMonitor.trigger.query },

                actions: codeMonitor.actions.nodes.map(action => ({
                    email: {
                        enabled: action.enabled,
                        priority: MonitorEmailPriority.NORMAL,
                        recipients: [props.authenticatedUser.id],
                        header: '',
                    },
                })),
            }),
        [props.authenticatedUser.id]
    )

    return (
        <div className="container col-8 mt-5">
            <PageTitle title="Create new code monitor" />
            <div className="page-header d-flex flex-wrap align-items-center">
                <h2 className="flex-grow-1">Create code monitor</h2>
            </div>
            Code monitors watch your code for specific triggers and run actions in response.{' '}
            <a href="" target="_blank" rel="noopener">
                {/* TODO: populate link */}
                Learn more
            </a>
            <CodeMonitorForm {...props} onSubmit={createMonitorRequest} submitButtonLabel="Create code monitor" />
        </div>
    )
}
