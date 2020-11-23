import * as H from 'history'
import classnames from 'classnames'
import React, { useCallback, useState } from 'react'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'
import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Link } from '../../../../shared/src/components/Link'
import { CodeMonitoringProps } from '.'
import PlusIcon from 'mdi-react/PlusIcon'
import { CodeMonitorNode } from './CodeMonitoringNode'

export interface CodeMonitoringPageProps extends BreadcrumbsProps, BreadcrumbSetters, CodeMonitoringProps {
    authenticatedUser: AuthenticatedUser
    location: H.Location
    history: H.History
}

type CodeMonitorFilter = 'all' | 'user'

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

    const [monitorListFilter, setMonitorListFilter] = useState<CodeMonitorFilter>('all')

    const setAllFilter = useCallback<React.MouseEventHandler>(() => {
        setMonitorListFilter('all')
    }, [])

    const setUserFilter = useCallback<React.MouseEventHandler>(() => {
        setMonitorListFilter('user')
    }, [])

    return (
        <div className="container mt-3 web-content">
            <PageTitle title="Code Monitoring" />
            <PageHeader
                title={
                    <>
                        Code monitoring{' '}
                        <sup>
                            <span className="badge badge-info text-uppercase">Prototype</span>
                        </sup>
                    </>
                }
                icon={VideoInputAntennaIcon}
                actions={
                    <Link to="/code-monitoring/new" className="btn btn-secondary">
                        <PlusIcon className="icon-inline" />
                        Add new code monitor
                    </Link>
                }
            />
            <div className="text-muted mb-4">
                {/* TODO: Add link to docs */}
                Watch your code for changes and trigger actions to get notifications, send webhooks, and more.{' '}
                <a href="/">Learn more.</a>
            </div>
            <div className="d-flex flex-column">
                <div className="code-monitoring-page-tabs border-bottom mb-4">
                    <div className="nav nav-tabs border-bottom-0">
                        <div className="nav-item">
                            <div className="nav-link active">Code monitors</div>
                        </div>
                    </div>
                </div>
                <div className="row mb-5">
                    <div className="d-flex flex-column col-2 mr-2">
                        <h3>Filters</h3>
                        <button
                            type="button"
                            className={classnames('btn text-left', { 'btn-primary': monitorListFilter === 'all' })}
                            onClick={setAllFilter}
                        >
                            All
                        </button>
                        <button
                            type="button"
                            className={classnames('btn text-left', { 'btn-primary': monitorListFilter === 'user' })}
                            onClick={setUserFilter}
                        >
                            Your code monitors
                        </button>
                    </div>
                    <div className="d-flex flex-column w-100 col">
                        <h3 className="mb-2">
                            {`${monitorListFilter === 'all' ? 'All code monitors' : 'Your code monitors'}`}
                        </h3>
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
                <div className="mt-5">
                    {/* TODO: add link */}
                    We want to hear your feedback! <a href="/">Share your thoughts</a>
                </div>
            </div>
        </div>
    )
}
