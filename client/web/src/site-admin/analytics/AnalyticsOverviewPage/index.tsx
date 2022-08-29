import React, { useState, useEffect } from 'react'

import {
    mdiCheck,
    mdiClose,
    mdiAccount,
    mdiSourceRepository,
    mdiCommentOutline,
    mdiArrowRight,
    mdiMagnify,
    mdiSitemap,
    mdiBookOutline,
    mdiPuzzleOutline,
} from '@mdi/js'
import classNames from 'classnames'
import format from 'date-fns/format'
import * as H from 'history'

import { useQuery } from '@sourcegraph/http-client'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Card, H2, H3, H4, Text, LoadingSpinner, Link, AnchorLink, Icon } from '@sourcegraph/wildcard'

import { BatchChangesIconNav } from '../../../batches/icons'
import { dismissAlert, isAlertDismissed } from '../../../components/DismissibleAlert'
import { OverviewStatisticsResult, OverviewStatisticsVariables, AnalyticsDateRange } from '../../../graphql-operations'
import { formatRelativeExpirationDate, isProductLicenseExpired } from '../../../productSubscription/helpers'
import { eventLogger } from '../../../tracking/eventLogger'
import { AnalyticsPageTitle } from '../components/AnalyticsPageTitle'
import { HorizontalSelect } from '../components/HorizontalSelect'
import { ValueLegendItem } from '../components/ValueLegendList'
import { useChartFilters } from '../useChartFilters'
import { formatNumber } from '../utils'

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
        {
            variables: {
                dateRange: dateRange.value,
            },
        }
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

    const { analytics, productSubscription } = data.site
    const licenseExpiresAt = productSubscription.license ? new Date(productSubscription.license.expiresAt) : null

    const totalSearchEvents =
        analytics.search.searches.summary.totalCount + analytics.search.fileViews.summary.totalCount

    const totalSearchHoursSaved =
        (totalSearchEvents * 0.75 * 0.5 + totalSearchEvents * 0.22 * 5 + totalSearchEvents * 0.03 * 120) / 60

    const totalCodeIntelEvents =
        analytics.codeIntel.definitionClicks.summary.totalCount + analytics.codeIntel.referenceClicks.summary.totalCount
    const totalCodeIntelHoverEvents =
        analytics.codeIntel.searchBasedEvents.summary.totalCount + analytics.codeIntel.preciseEvents.summary.totalCount

    const totalCodeIntelHoursSaved =
        (analytics.codeIntel.inAppEvents.summary.totalCount * 0.5 +
            analytics.codeIntel.codeHostEvents.summary.totalCount * 1.5 +
            Math.floor(
                (analytics.codeIntel.crossRepoEvents.summary.totalCount * totalCodeIntelEvents * 3) /
                    totalCodeIntelHoverEvents || 0
            ) +
            Math.floor(
                (analytics.codeIntel.preciseEvents.summary.totalCount * totalCodeIntelEvents) /
                    totalCodeIntelHoverEvents || 0
            )) /
        60

    const totalBatchChangesEvents = analytics.batchChanges.changesetsMerged.summary.totalCount
    const totalBatchChangesHoursSaved = (totalBatchChangesEvents * 15) / 60

    const totalNotebooksEvents = analytics.notebooks.views.summary.totalCount
    const totalNotebooksHoursSaved = (totalNotebooksEvents * 5) / 60

    const totalExtensionsEvents =
        analytics.extensions.vscode.summary.totalCount +
        analytics.extensions.jetbrains.summary.totalCount +
        analytics.extensions.browser.summary.totalCount

    const totalExtensionsHoursSaved =
        (analytics.extensions.vscode.summary.totalCount * 3 +
            analytics.extensions.jetbrains.summary.totalCount * 1.5 +
            analytics.extensions.browser.summary.totalCount * 0.5) /
        60

    const totalEvents =
        totalSearchEvents +
        totalCodeIntelEvents +
        totalBatchChangesEvents +
        totalNotebooksEvents +
        totalExtensionsEvents

    const totalHoursSaved =
        totalSearchHoursSaved +
        totalCodeIntelHoursSaved +
        totalBatchChangesHoursSaved +
        totalNotebooksHoursSaved +
        totalExtensionsHoursSaved

    const projectedHoursSaved = (() => {
        if (dateRange.value === AnalyticsDateRange.LAST_WEEK) {
            return totalHoursSaved * 52
        }
        if (dateRange.value === AnalyticsDateRange.LAST_MONTH) {
            return totalHoursSaved * 12
        }
        if (dateRange.value === AnalyticsDateRange.LAST_THREE_MONTHS) {
            return (totalHoursSaved * 12) / 3
        }
        return totalHoursSaved
    })()

    return (
        <>
            <AnalyticsPageTitle>Overview</AnalyticsPageTitle>

            <Card className="p-3">
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
                    <div className={classNames('my-3', styles.padded)}>
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
                        <H3 className="mb-3">Developer time saved</H3>
                        <div className={classNames(styles.statsBox, 'p-4 mb-3')}>
                            <div className="d-flex">
                                <ValueLegendItem
                                    value={data.site.analytics.users.activity.summary.totalUniqueUsers}
                                    className={classNames('flex-1', styles.borderRight)}
                                    description="Active Users"
                                    color="var(--body-color)"
                                    tooltip="Users using the application in the selected timeframe."
                                />
                                <ValueLegendItem
                                    value={totalEvents}
                                    className={classNames('flex-1', styles.borderRight)}
                                    description="Events"
                                    color="var(--body-color)"
                                    tooltip="Total number of actions performed in the selected timeframe."
                                />
                                <ValueLegendItem
                                    value={totalHoursSaved}
                                    className="flex-1"
                                    description="Hours saved"
                                    color="var(--purple)"
                                    tooltip="Total number of hours saved in the selected timeframe."
                                />
                            </div>
                            {data.users.totalCount > 1 && (
                                <div className="d-flex flex-column align-items-center mt-4">
                                    <H2>
                                        Annual projection:{' '}
                                        <span className={styles.purple}>{formatNumber(projectedHoursSaved)} hours</span>{' '}
                                        saved*
                                    </H2>
                                    <Text as="span" className="text-muted">
                                        * Based on{' '}
                                        {dateRange.value === AnalyticsDateRange.LAST_THREE_MONTHS
                                            ? 'last 3 months'
                                            : dateRange.value === AnalyticsDateRange.LAST_MONTH
                                            ? 'last month'
                                            : 'last week'}{' '}
                                        of data
                                    </Text>
                                </div>
                            )}
                        </div>
                        <H3 className={classNames('my-3 pb-2', styles.border)}>Hours by feature</H3>
                        <table className={styles.hoursTable}>
                            <thead>
                                <tr>
                                    <Text as="th" className="text-muted text-left">
                                        EVENT TYPE
                                    </Text>
                                    <Text as="th" className="text-muted">
                                        EVENTS
                                    </Text>
                                    <Text as="th" className="text-muted">
                                        HOURS SAVED
                                    </Text>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td className="text-left">
                                        <Link to="/site-admin/analytics/search">
                                            <Text as="span" className="d-flex align-items-center">
                                                <Icon
                                                    svgPath={mdiMagnify}
                                                    size="md"
                                                    aria-label="Code Search"
                                                    className="mr-1"
                                                />
                                                Search
                                            </Text>
                                        </Link>
                                    </td>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalSearchEvents)}
                                    </Text>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalSearchHoursSaved)}
                                    </Text>
                                </tr>
                                <tr>
                                    <td className="text-left">
                                        <Link to="/site-admin/analytics/code-intel">
                                            <Text as="span" className="d-flex align-items-center">
                                                <Icon
                                                    svgPath={mdiSitemap}
                                                    size="md"
                                                    aria-label="Code Navigation"
                                                    className="mr-1"
                                                />
                                                Code Navigation
                                            </Text>
                                        </Link>
                                    </td>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalCodeIntelEvents)}
                                    </Text>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalCodeIntelHoursSaved)}
                                    </Text>
                                </tr>
                                <tr>
                                    <td className="text-left">
                                        <Link to="/site-admin/analytics/batch-changes">
                                            <Text as="span" className="d-flex align-items-center">
                                                <BatchChangesIconNav className="mr-1" />
                                                Batch Changes
                                            </Text>
                                        </Link>
                                    </td>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalBatchChangesEvents)}
                                    </Text>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalBatchChangesHoursSaved)}
                                    </Text>
                                </tr>
                                <tr>
                                    <td className="text-left">
                                        <Link to="/site-admin/analytics/notebooks">
                                            <Text as="span" className="d-flex align-items-center">
                                                <Icon
                                                    svgPath={mdiBookOutline}
                                                    size="md"
                                                    aria-label="Notebooks"
                                                    className="mr-1"
                                                />
                                                Notebooks
                                            </Text>
                                        </Link>
                                    </td>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalNotebooksEvents)}
                                    </Text>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalNotebooksHoursSaved)}
                                    </Text>
                                </tr>
                                <tr>
                                    <td className="text-left">
                                        <Link to="/site-admin/analytics/extensions">
                                            <Text as="span" className="d-flex align-items-center">
                                                <Icon
                                                    svgPath={mdiPuzzleOutline}
                                                    size="md"
                                                    aria-label="Extensions"
                                                    className="mr-1"
                                                />
                                                Extensions
                                            </Text>
                                        </Link>
                                    </td>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalExtensionsEvents)}
                                    </Text>
                                    <Text as="td" weight="bold">
                                        {formatNumber(totalExtensionsHoursSaved)}
                                    </Text>
                                </tr>
                            </tbody>
                        </table>
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
