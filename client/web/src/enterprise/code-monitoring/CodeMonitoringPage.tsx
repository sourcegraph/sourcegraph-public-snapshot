import * as H from 'history'
import React, { useCallback } from 'react'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'
import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { Link } from '../../../../shared/src/components/Link'
import { CodeMonitoringProps } from '.'

export interface CodeMonitoringPageProps extends BreadcrumbsProps, BreadcrumbSetters, CodeMonitoringProps {
    authenticatedUser: AuthenticatedUser
    location: H.Location
    history: H.History
}

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = props => {
    const { authenticatedUser, fetchUserCodeMonitors } = props

    const queryConnection = useCallback(
        (args: Partial<ListUserCodeMonitorsVariables>) =>
            fetchUserCodeMonitors({
                id: authenticatedUser.id,
                first: args.first ?? null,
                after: args.after ?? null,
            }),
        [authenticatedUser, fetchUserCodeMonitors]
    )

    return (
        <div className="container mt-3 web-content">
            <PageTitle title="Code Monitoring" />
            <PageHeader
                title={
                    <>
                        Code Monitoring{' '}
                        <sup>
                            <span className="badge badge-info text-uppercase">Prototype</span>
                        </sup>
                    </>
                }
                icon={VideoInputAntennaIcon}
            />
            <div className="d-flex flex-column">
                <div className="code-monitoring-page-tabs border-bottom mb-4">
                    <div className="nav nav-tabs border-bottom-0">
                        <div className="nav-item">
                            <div className="nav-link active">Code monitors</div>
                        </div>
                    </div>
                </div>
                <div>
                    <h3 className="mb-2">Your code monitors</h3>
                    <FilteredConnection<CodeMonitorFields>
                        location={props.location}
                        history={props.history}
                        defaultFirst={10}
                        queryConnection={queryConnection}
                        hideSearch={true}
                        nodeComponent={CodeMonitorNode}
                        noun="code monitor"
                        pluralNoun="code monitors"
                        noSummaryIfAllNodesVisible={true}
                        cursorPaging={true}
                    />
                </div>
            </div>
            <Link to="/code-monitoring/new">Add new code monitor</Link>
        </div>
    )
}

interface CodeMonitorNodeProps {
    node: CodeMonitorFields
}

const CodeMonitorNode: React.FunctionComponent<CodeMonitorNodeProps> = ({ node }: CodeMonitorNodeProps) => (
    <div className="card p-3 mb-2">
        <div className="d-flex justify-content-between align-items-center">
            <div className="d-flex flex-column">
                <div className="font-weight-bold">{node.description}</div>
                {node.actions.nodes.length > 0 && (
                    <div className="text-muted">New search result → Sends email notifications, delivers webhook</div>
                )}
            </div>
            <div>
                <Toggle value={node.enabled} />
            </div>
        </div>
    </div>
)
