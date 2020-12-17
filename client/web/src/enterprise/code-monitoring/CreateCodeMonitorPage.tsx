import * as H from 'history'
import React, { useCallback, useMemo } from 'react'
import { Observable } from 'rxjs'
import { CodeMonitoringProps } from '.'
import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorFields, MonitorEmailPriority } from '../../graphql-operations'
import { CodeMonitorForm } from './components/CodeMonitorForm'

export interface CreateCodeMonitorPageProps
    extends BreadcrumbsProps,
        BreadcrumbSetters,
        Pick<CodeMonitoringProps, 'createCodeMonitor'> {
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

    const { authenticatedUser, createCodeMonitor } = props
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
        <div className="container col-8 mt-5">
            <PageTitle title="Create new code monitor" />
            <div className="page-header d-flex flex-wrap align-items-center">
                <h2 className="flex-grow-1">Create code monitor</h2>
            </div>
            Code monitors watch your code for specific triggers and run actions in response.{' '}
            <a
                href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                target="_blank"
                rel="noopener"
            >
                Learn more
            </a>
            <CodeMonitorForm {...props} onSubmit={createMonitorRequest} submitButtonLabel="Create code monitor" />
        </div>
    )
}
