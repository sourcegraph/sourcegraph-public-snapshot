import * as H from 'history'
import classnames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'
import { AuthenticatedUser } from '../../auth'
import { FilteredConnection } from '../../components/FilteredConnection'
import { CodeMonitorFields, ListUserCodeMonitorsResult, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { Link } from '../../../../shared/src/components/Link'
import { CodeMonitoringProps } from '.'
import PlusIcon from 'mdi-react/PlusIcon'
import { CodeMonitorNode, CodeMonitorNodeProps } from './CodeMonitoringNode'
import { catchError, map, startWith, tap } from 'rxjs/operators'
import { asError, isErrorLike } from '../../../../shared/src/util/errors'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export interface CodeMonitoringPageProps
    extends BreadcrumbsProps,
        BreadcrumbSetters,
        Pick<CodeMonitoringProps, 'fetchUserCodeMonitors' | 'toggleCodeMonitorEnabled'> {
    authenticatedUser: AuthenticatedUser
    location: H.Location
    history: H.History
}

type CodeMonitorFilter = 'all' | 'user'

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = props => {
    const { authenticatedUser, fetchUserCodeMonitors, toggleCodeMonitorEnabled } = props

    const queryConnection = useCallback(
        (args: Partial<ListUserCodeMonitorsVariables>) =>
            fetchUserCodeMonitors({
                id: authenticatedUser.id,
                first: args.first ?? null,
                after: args.after ?? null,
            }),
        [authenticatedUser, fetchUserCodeMonitors]
    )

    const LOADING = 'loading' as const

    const userHasCodeMonitors = useObservable(
        useMemo(
            () =>
                fetchUserCodeMonitors({
                    id: authenticatedUser.id,
                    first: 1,
                    after: null,
                }).pipe(
                    map(monitors => {
                        if (monitors.nodes.length > 0) {
                            return true
                        }
                        return false
                    }),
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [authenticatedUser.id, fetchUserCodeMonitors]
        )
    )

    const [monitorListFilter, setMonitorListFilter] = useState<CodeMonitorFilter>('all')

    const setAllFilter = useCallback<React.MouseEventHandler>(() => {
        setMonitorListFilter('all')
    }, [])

    const setUserFilter = useCallback<React.MouseEventHandler>(() => {
        setMonitorListFilter('user')
    }, [])

    return (
        <div className="container mt-5">
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
                    userHasCodeMonitors &&
                    userHasCodeMonitors !== 'loading' &&
                    !isErrorLike(userHasCodeMonitors) && (
                        <Link to="/code-monitoring/new" className="btn btn-secondary">
                            <PlusIcon className="icon-inline" />
                            Create new code monitor
                        </Link>
                    )
                }
            />
            {userHasCodeMonitors === 'loading' && <LoadingSpinner />}
            {!userHasCodeMonitors && (
                <div className="mt-5">
                    <div className="d-flex flex-column mb-5">
                        <h2>Get started with code monitoring</h2>
                        <p className="text-muted code-monitoring-page__start-subheading mb-4">
                            Watch your code for changes and trigger actions to get notifications, send webhooks, and
                            more. <a href="">Learn more.</a>
                        </p>
                        <Link
                            to="/code-monitoring/new"
                            className="code-monitoring-page__start-button btn btn-primary"
                            type="button"
                        >
                            Create your first code monitor →
                        </Link>
                    </div>
                    <div className="code-monitoring-page__start-points container">
                        <h3>Starting points for your first monitor</h3>
                        <div className="row no-gutters code-monitoring-page__start-points-panel-container">
                            <div className="col-6">
                                <div className="code-monitoring-page__start-points-panel">
                                    <h3>Watch for AWS secrets in commits</h3>
                                    <p className="text-muted">
                                        Use a search query to watch for new search results, and choose how to receive
                                        notifications in response.
                                    </p>
                                </div>
                            </div>
                            <div className="col-6">
                                <div className="code-monitoring-page__start-points-panel">
                                    <h3>Watch for new consumers of deprecated methods</h3>
                                    <p className="text-muted">
                                        Keep an eye on commits with new consumers of deprecated methods to keep your
                                        code base up-to-date.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
            {userHasCodeMonitors && userHasCodeMonitors !== 'loading' && !isErrorLike(userHasCodeMonitors) && (
                <>
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
                                    className={classnames('btn text-left', {
                                        'btn-primary': monitorListFilter === 'all',
                                    })}
                                    onClick={setAllFilter}
                                >
                                    All
                                </button>
                                <button
                                    type="button"
                                    className={classnames('btn text-left', {
                                        'btn-primary': monitorListFilter === 'user',
                                    })}
                                    onClick={setUserFilter}
                                >
                                    Your code monitors
                                </button>
                            </div>
                            <div className="d-flex flex-column w-100 col">
                                <h3 className="mb-2">
                                    {`${monitorListFilter === 'all' ? 'All code monitors' : 'Your code monitors'}`}
                                </h3>
                                <FilteredConnection<
                                    CodeMonitorFields,
                                    Omit<CodeMonitorNodeProps, 'node'>,
                                    (ListUserCodeMonitorsResult['node'] & { __typename: 'User' })['monitors']
                                >
                                    location={props.location}
                                    history={props.history}
                                    defaultFirst={10}
                                    queryConnection={queryConnection}
                                    hideSearch={true}
                                    nodeComponent={CodeMonitorNode}
                                    nodeComponentProps={{
                                        location: props.location,
                                        toggleCodeMonitorEnabled,
                                    }}
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
                </>
            )}
        </div>
    )
}
