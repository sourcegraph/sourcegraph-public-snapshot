import React, { useState, useEffect } from 'react'

import { mdiCheck, mdiClose, mdiAccount, mdiSourceRepository, mdiCommentOutline, mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'
import format from 'date-fns/format'
import * as H from 'history'

import { useQuery } from '@sourcegraph/http-client'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Card, H2, H3, H4, Text, LoadingSpinner, AnchorLink, Icon } from '@sourcegraph/wildcard'

import { dismissAlert, isAlertDismissed } from '../../../components/DismissibleAlert'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { OverviewStatisticsResult, OverviewStatisticsVariables } from '../../../graphql-operations'
import { formatRelativeExpirationDate, isProductLicenseExpired } from '../../../productSubscription/helpers'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { useChartFilters } from '../useChartFilters'

import { DevTimeSaved } from './DevTimeSaved'
import { OVERVIEW_STATISTICS } from './queries'
import { Sidebar } from './Sidebar'

import styles from './index.module.scss'

const GET_STARTED_ALERT_KEY = 'get_started_admin_analytics_overview'

interface IProps extends ActivationProps {
    history: H.History
}

export const AnalyticsOverviewPage: React.FunctionComponent<IProps> = ({ activation, history }) => {
    const { dateRange } = useChartFilters({ name: 'Overview' })
    const [showGetStarted, setShowGetStarted] = useState(!isAlertDismissed(GET_STARTED_ALERT_KEY))
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
                                        to="https://about.sourcegraph.com/pricing"
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
                {showGetStarted && activation && (
                    <div className={classNames('my-3', styles.padded)} data-testid="site-admin-overview-menu">
                        <div className={styles.getStartedBox}>
                            <div className="d-flex justify-content-between align-items-center">
                                <H3>Get started with Sourcegraph</H3>
                                <Icon
                                    svgPath={mdiClose}
                                    aria-label="Close Get started alert"
                                    className="cursor-pointer"
                                    onClick={() => {
                                        dismissAlert(GET_STARTED_ALERT_KEY)
                                        setShowGetStarted(false)
                                    }}
                                />
                            </div>

                            {activation.steps.map((step, index) => (
                                <div
                                    key={step.id}
                                    onClick={event => step.onClick?.(event, history)}
                                    onKeyDown={() => undefined}
                                    role="button"
                                    tabIndex={0}
                                    className={classNames('d-flex py-3 align-items-center', {
                                        [styles.borderTop]: index > 0,
                                        'cursor-pointer': !!step.onClick,
                                    })}
                                >
                                    <div
                                        className={classNames(styles.getStartedTick, {
                                            [styles.completed]: activation?.completed?.[step.id],
                                        })}
                                    >
                                        <Icon
                                            svgPath={mdiCheck}
                                            size="md"
                                            aria-label="Get started item completed status"
                                        />
                                    </div>
                                    <div className="mx-3 flex-1">
                                        <H4 className={classNames(styles.link, 'mb-0')}>{step.title}</H4>
                                        <Text className="mb-0">{step.detail}</Text>
                                    </div>
                                    <Icon
                                        svgPath={mdiArrowRight}
                                        size="md"
                                        aria-label="Get started item link"
                                        className={styles.getStartedLink}
                                    />
                                </div>
                            ))}
                        </div>
                    </div>
                )}
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
                                            label: 'Bytes stored',
                                            value: Number(data.repositoryStats.gitDirBytes),
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
