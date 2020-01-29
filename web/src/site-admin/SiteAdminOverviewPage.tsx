import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Observable, of } from 'rxjs'
import { map, catchError } from 'rxjs/operators'
import { ActivationProps, percentageDone } from '../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../shared/src/components/activation/ActivationChecklist'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { numberWithCommas, pluralize } from '../../../shared/src/util/strings'
import { queryGraphQL } from '../backend/graphql'
import { Collapsible } from '../components/Collapsible'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { UsageChart } from './SiteAdminUsageStatisticsPage'
import { ErrorAlert } from '../components/alerts'
import { useObservable } from '../util/useObservable'
import { ErrorLike, asError, isErrorLike } from '../../../shared/src/util/errors'

interface Props extends ActivationProps {
    history: H.History
    overviewComponents: readonly React.ComponentType[]
    isLightTheme: boolean
}

const fetchOverview = (): Observable<{
    repositories: number | null
    users: number
    orgs: number
    surveyResponses: {
        totalCount: number
        averageScore: number
    }
}> =>
    queryGraphQL(gql`
        query Overview {
            repositories {
                totalCount(precise: true)
            }
            users {
                totalCount
            }
            organizations {
                totalCount
            }
            surveyResponses {
                totalCount
                averageScore
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => ({
            repositories: data.repositories.totalCount,
            users: data.users.totalCount,
            orgs: data.organizations.totalCount,
            surveyResponses: data.surveyResponses,
        }))
    )

const fetchWeeklyActiveUsers = (): Observable<GQL.ISiteUsageStatistics> =>
    queryGraphQL(gql`
        query WAUs {
            site {
                usageStatistics {
                    waus {
                        userCount
                        registeredUserCount
                        anonymousUserCount
                        startTime
                    }
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.usageStatistics)
    )

/**
 * A page displaying an overview of site admin information.
 */
export const SiteAdminOverviewPage: React.FunctionComponent<Props> = props => {
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminOverview')
    }, [])

    const info = useObservable(
        useMemo(() => fetchOverview().pipe(catchError(error => of<ErrorLike>(asError(error)))), [])
    )

    const stats = useObservable(
        useMemo(() => fetchWeeklyActiveUsers().pipe(catchError(error => of<ErrorLike>(asError(error)))), [])
    )

    let setupPercentage = 0
    if (props.activation) {
        setupPercentage = percentageDone(props.activation.completed)
    }
    return (
        <div className="site-admin-overview-page">
            <PageTitle title="Overview - Admin" />
            {props.overviewComponents.length > 0 && (
                <div className="mb-4">
                    {props.overviewComponents.map((C, i) => (
                        <C key={i} />
                    ))}
                </div>
            )}
            {info === undefined && <LoadingSpinner className="icon-inline" />}
            <div className="list-group">
                {info && !isErrorLike(info) && (
                    <>
                        {props.activation && props.activation.completed && (
                            <Collapsible
                                title={<>{setupPercentage < 100 ? 'Get started with Sourcegraph' : 'Setup status'}</>}
                                defaultExpanded={setupPercentage < 100}
                                className="list-group-item"
                                titleClassName="h4 mb-0 mt-2 font-weight-normal p-2"
                            >
                                {props.activation.completed && (
                                    <ActivationChecklist
                                        history={props.history}
                                        steps={props.activation.steps}
                                        completed={props.activation.completed}
                                    />
                                )}
                            </Collapsible>
                        )}
                        {info.repositories !== null && (
                            <Link
                                to="/site-admin/repositories"
                                className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                            >
                                {numberWithCommas(info.repositories)}{' '}
                                {pluralize('repository', info.repositories, 'repositories')}
                            </Link>
                        )}
                        {info.users > 1 && (
                            <Link
                                to="/site-admin/users"
                                className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                            >
                                {numberWithCommas(info.users)} {pluralize('user', info.users)}
                            </Link>
                        )}
                        {info.orgs > 1 && (
                            <Link
                                to="/site-admin/organizations"
                                className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                            >
                                {numberWithCommas(info.orgs)} {pluralize('organization', info.orgs)}
                            </Link>
                        )}
                        {info.users > 1 && (
                            <Link
                                to="/site-admin/surveys"
                                className="list-group-item list-group-item-action h5 font-weight-normal py-2 px-3"
                            >
                                {numberWithCommas(info.surveyResponses.totalCount)}{' '}
                                {pluralize('user survey response', info.surveyResponses.totalCount)}
                            </Link>
                        )}
                        {info.users > 1 &&
                            stats !== undefined &&
                            (isErrorLike(stats) ? (
                                <ErrorAlert className="mb-3" error={stats} />
                            ) : (
                                <Collapsible
                                    title={
                                        <>
                                            {stats.waus[1].userCount}{' '}
                                            {pluralize('active user', stats.waus[1].userCount)} last week
                                        </>
                                    }
                                    defaultExpanded={true}
                                    className="list-group-item"
                                    titleClassName="h5 mb-0 font-weight-normal p-2"
                                >
                                    {stats && (
                                        <UsageChart
                                            {...props}
                                            stats={stats}
                                            chartID="waus"
                                            showLegend={false}
                                            header={
                                                <div className="site-admin-overview-page__detail-header">
                                                    <h2>Weekly unique users</h2>
                                                    <h3>
                                                        <Link
                                                            to="/site-admin/usage-statistics"
                                                            className="btn btn-secondary"
                                                        >
                                                            View all usage statistics{' '}
                                                            <OpenInNewIcon className="icon-inline" />
                                                        </Link>
                                                    </h3>
                                                </div>
                                            }
                                        />
                                    )}
                                </Collapsible>
                            ))}
                    </>
                )}
            </div>
        </div>
    )
}
