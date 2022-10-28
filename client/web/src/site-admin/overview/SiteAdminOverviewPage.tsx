import React, { useEffect, useMemo } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { Observable, of } from 'rxjs'
import { map, catchError } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, asError, isErrorLike, numberWithCommas, pluralize } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { LoadingSpinner, useObservable, Button, Link, Icon, H3, Heading } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { Collapsible } from '../../components/Collapsible'
import { PageTitle } from '../../components/PageTitle'
import { Scalars, WAUsResult } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { SiteUsagePeriodFragment } from '../backend'
import { UsageChart } from '../SiteAdminUsageStatisticsPage'

import styles from './SiteAdminOverviewPage.module.scss'

interface Props extends ThemeProps {
    overviewComponents: readonly React.ComponentType<React.PropsWithChildren<unknown>>[]

    /** For testing only */
    _fetchOverview?: () => Observable<{
        repositories: number | null
        repositoryStats: {
            gitDirBytes: Scalars['BigInt']
            indexedLinesCount: Scalars['BigInt']
        }
        users: number
        orgs: number
        surveyResponses: {
            totalCount: number
            averageScore: number
        }
    }>
    /** For testing only */
    _fetchWeeklyActiveUsers?: () => Observable<WAUsResult['site']['usageStatistics']>
}

const fetchOverview = (): Observable<{
    repositories: number | null
    repositoryStats: {
        gitDirBytes: Scalars['BigInt']
        indexedLinesCount: Scalars['BigInt']
    }
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
            repositoryStats {
                gitDirBytes
                indexedLinesCount
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
            repositoryStats: data.repositoryStats,
            users: data.users.totalCount,
            orgs: data.organizations.totalCount,
            surveyResponses: data.surveyResponses,
        }))
    )

const fetchWeeklyActiveUsers = (): Observable<WAUsResult['site']['usageStatistics']> =>
    queryGraphQL(gql`
        query WAUs {
            site {
                usageStatistics {
                    waus {
                        ...SiteUsagePeriodFields
                    }
                }
            }
        }

        ${SiteUsagePeriodFragment}
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.usageStatistics)
    )

/**
 * A page displaying an overview of site admin information.
 */
export const SiteAdminOverviewPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isLightTheme,
    overviewComponents,
    _fetchOverview = fetchOverview,
    _fetchWeeklyActiveUsers = fetchWeeklyActiveUsers,
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminOverview')
    }, [])

    const info = useObservable(
        useMemo(() => _fetchOverview().pipe(catchError(error => of<ErrorLike>(asError(error)))), [_fetchOverview])
    )

    const stats = useObservable(
        useMemo(
            () =>
                _fetchWeeklyActiveUsers().pipe(
                    map(({ waus }) => ({ waus, daus: [], maus: [] })),
                    catchError(error => of<ErrorLike>(asError(error)))
                ),
            [_fetchWeeklyActiveUsers]
        )
    )

    return (
        <div className="site-admin-overview-page">
            <PageTitle title="Overview - Admin" />
            {overviewComponents.length > 0 && (
                <div className="mb-4">
                    {overviewComponents.map((Component, index) => (
                        <Component key={index} />
                    ))}
                </div>
            )}
            {info === undefined && <LoadingSpinner />}
            <div className="list-group">
                {info && !isErrorLike(info) && (
                    <>
                        {info.repositories !== null && (
                            <Link
                                to="/site-admin/repositories"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
                            >
                                {numberWithCommas(info.repositories)}{' '}
                                {pluralize('repository', info.repositories, 'repositories')}
                            </Link>
                        )}
                        {info.repositoryStats !== null && (
                            <Link
                                to="/site-admin/repositories"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
                            >
                                {BigInt(info.repositoryStats.gitDirBytes).toLocaleString()}{' '}
                                {pluralize('byte stored', BigInt(info.repositoryStats.gitDirBytes), 'bytes stored')}
                            </Link>
                        )}
                        {info.repositoryStats !== null && (
                            <Link
                                to="/site-admin/repositories"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
                            >
                                {BigInt(info.repositoryStats.indexedLinesCount).toLocaleString()}{' '}
                                {pluralize(
                                    'line of code indexed',
                                    BigInt(info.repositoryStats.indexedLinesCount),
                                    'lines of code indexed'
                                )}
                            </Link>
                        )}
                        {info.users > 1 && (
                            <Link
                                to="/site-admin/users"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
                            >
                                {numberWithCommas(info.users)} {pluralize('user', info.users)}
                            </Link>
                        )}
                        {info.orgs > 1 && (
                            <Link
                                to="/site-admin/organizations"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
                            >
                                {numberWithCommas(info.orgs)} {pluralize('organization', info.orgs)}
                            </Link>
                        )}
                        {info.users > 1 && (
                            <Link
                                to="/site-admin/surveys"
                                className={classNames(
                                    'list-group-item list-group-item-action mb-0 font-weight-normal py-2 px-3',
                                    styles.adminOverviewMenuText
                                )}
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
                                    titleClassName={classNames(
                                        'mb-0 font-weight-normal p-2',
                                        styles.adminOverviewMenuText
                                    )}
                                    titleAtStart={true}
                                >
                                    {stats && (
                                        <UsageChart
                                            isLightTheme={isLightTheme}
                                            stats={stats}
                                            chartID="waus"
                                            showLegend={false}
                                            header={
                                                <div className="site-admin-overview-page__detail-header">
                                                    <Heading as="h3" styleAs="h2">
                                                        Weekly unique users
                                                    </Heading>
                                                    <H3>
                                                        <Button
                                                            to="/site-admin/usage-statistics"
                                                            variant="secondary"
                                                            as={Link}
                                                        >
                                                            View all usage statistics{' '}
                                                            <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                                                        </Button>
                                                    </H3>
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
