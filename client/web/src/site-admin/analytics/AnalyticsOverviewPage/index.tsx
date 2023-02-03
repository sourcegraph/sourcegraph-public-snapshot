import React, { useEffect } from 'react'

import { mdiAccount, mdiSourceRepository, mdiCommentOutline } from '@mdi/js'
import classNames from 'classnames'
import format from 'date-fns/format'
import * as H from 'history'

import { useQuery } from '@sourcegraph/http-client'
import { Card, H2, Text, LoadingSpinner, AnchorLink } from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { OverviewStatisticsResult, OverviewStatisticsVariables } from '../../../graphql-operations'
import { formatRelativeExpirationDate, isProductLicenseExpired } from '../../../productSubscription/helpers'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { useChartFilters } from '../useChartFilters'
import { getByteUnitLabel, getByteUnitValue } from '../utils'

import { DevTimeSaved } from './DevTimeSaved'
import { OVERVIEW_STATISTICS } from './queries'
import { Sidebar } from './Sidebar'

import styles from './index.module.scss'

interface IProps {
    history: H.History
}

export const AnalyticsOverviewPage: React.FunctionComponent<IProps> = ({ history }) => {
    const { dateRange } = useChartFilters({ name: 'Overview' })
    const { data, error, loading } = useQuery<OverviewStatisticsResult, OverviewStatisticsVariables>(
        OVERVIEW_STATISTICS,
        {}
    )
    useEffect(() => {
        eventLogger.logPageView('AdminAnalyticsOverview')
    }, [])

    if (error) {
        throw error
    }

    if (loading || !data) {
        return <LoadingSpinner />
    }

    const { productSubscription } = data.site
    const licenseExpiresAt = productSubscription.license ? new Date(productSubscription.license.expiresAt) : null

    return (
        <>
            <AnalyticsPageTitle>Overview</AnalyticsPageTitle>

            <Card className="p-3" data-testid="product-certificate">
                <div className="d-flex justify-content-between align-items-start mb-3 text-nowrap">
                    <div>
                        <H2 className="mb-3">{data.site.productSubscription.productNameWithBrand}</H2>
                        <div className="d-flex">
                            <Text className="text-muted">
                                Version <span className={styles.purple}>{data.site.productVersion}</span>
                            </Text>
                            {productSubscription.license && licenseExpiresAt ? (
                                <>
                                    <AnchorLink
                                        to="/help/admin/updates"
                                        target="_blank"
                                        rel="noopener"
                                        className="ml-1"
                                    >
                                        Upgrade
                                    </AnchorLink>
                                    <Text className="text-muted mx-2">|</Text>
                                    <Text className="text-muted">
                                        License
                                        {isProductLicenseExpired(licenseExpiresAt) ? ' expired on ' : ' valid until '}
                                        <span title={format(licenseExpiresAt, 'PPpp')}>
                                            {format(licenseExpiresAt, 'yyyy-MM-dd')}
                                        </span>{' '}
                                        ({formatRelativeExpirationDate(licenseExpiresAt)})
                                    </Text>
                                </>
                            ) : (
                                <AnchorLink
                                    to="http://about.sourcegraph.com/contact/sales"
                                    target="_blank"
                                    rel="noopener"
                                    className="ml-1"
                                >
                                    Get license
                                </AnchorLink>
                            )}
                        </div>
                    </div>
                    <HorizontalSelect<typeof dateRange.value> {...dateRange} />
                </div>
                <div className={classNames('d-flex mt-3', styles.padded)}>
                    <div className={styles.main}>
                        <ErrorBoundary location={null}>
                            <DevTimeSaved
                                showAnnualProjection={data.users.totalCount > 1}
                                dateRange={dateRange.value}
                            />
                        </ErrorBoundary>
                    </div>
                    <div className={styles.sidebar}>
                        <Sidebar
                            sections={[
                                {
                                    title: 'Users statistics',
                                    icon: mdiAccount,
                                    link: '/site-admin/analytics/users',
                                    items: [
                                        { label: 'Total users', value: data.users.totalCount },
                                        {
                                            label: 'Administrators',
                                            value: data.site.adminUsers.totalCount,
                                        },
                                        {
                                            label: 'Users licenses',
                                            value: data.site.productSubscription.license?.userCount || 0,
                                        },
                                    ],
                                },
                                {
                                    title: 'Code statistics',
                                    icon: mdiSourceRepository,
                                    link: '/site-admin/repositories',
                                    items: [
                                        {
                                            label: 'Repositories',
                                            value: data.repositories.totalCount || 0,
                                        },
                                        {
                                            label: `${getByteUnitLabel(
                                                Number(data.repositoryStats.gitDirBytes)
                                            )} stored`,
                                            value: getByteUnitValue(Number(data.repositoryStats.gitDirBytes)),
                                        },
                                        {
                                            label: 'Lines of code',
                                            value: Number(data.repositoryStats.indexedLinesCount),
                                        },
                                    ],
                                },
                                {
                                    title: 'Feedback',
                                    icon: mdiCommentOutline,
                                    link: '/site-admin/surveys',
                                    items: [
                                        {
                                            label: 'Submissions',
                                            value: data.surveyResponses.totalCount,
                                        },
                                        {
                                            label: 'Avg score',
                                            value: data.surveyResponses.averageScore,
                                        },
                                        {
                                            label: 'NPS',
                                            value: data.surveyResponses.netPromoterScore,
                                        },
                                    ],
                                },
                            ]}
                        />
                    </div>
                </div>
            </Card>
        </>
    )
}
