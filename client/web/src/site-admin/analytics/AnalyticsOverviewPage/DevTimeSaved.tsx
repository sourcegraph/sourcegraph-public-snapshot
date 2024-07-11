import React from 'react'

import { mdiBookOutline, mdiMagnify, mdiPoll, mdiPuzzleOutline, mdiSitemap } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { H2, H3, Icon, Link, LoadingSpinner, Text, Tooltip } from '@sourcegraph/wildcard'

import { BatchChangesIconNav } from '../../../batches/icons'
import {
    AnalyticsDateRange,
    type OverviewDevTimeSavedResult,
    type OverviewDevTimeSavedVariables,
} from '../../../graphql-operations'
import { ValueLegendItem } from '../components/ValueLegendList'
import { formatNumber } from '../utils'

import { OVERVIEW_DEV_TIME_SAVED } from './queries'

import styles from './index.module.scss'

interface DevTimeSavedProps {
    showAnnualProjection?: boolean
    dateRange: AnalyticsDateRange
}

export const DevTimeSaved: React.FunctionComponent<DevTimeSavedProps> = ({ showAnnualProjection, dateRange }) => {
    const { data, error, loading } = useQuery<OverviewDevTimeSavedResult, OverviewDevTimeSavedVariables>(
        OVERVIEW_DEV_TIME_SAVED,
        {
            variables: {
                dateRange,
            },
        }
    )

    if (error) {
        throw error
    }

    if (loading || !data) {
        return (
            <div className="d-flex justify-content-center">
                <LoadingSpinner />
            </div>
        )
    }

    const { search, codeIntel, batchChanges, notebooks, extensions, users } = data.site.analytics

    const totalSearchEvents = search.searches.summary.totalCount + search.fileViews.summary.totalCount

    const totalSearchHoursSaved =
        (totalSearchEvents * 0.75 * 0.5 + totalSearchEvents * 0.22 * 5 + totalSearchEvents * 0.03 * 120) / 60

    const totalCodeIntelEvents =
        codeIntel.definitionClicks.summary.totalCount + codeIntel.referenceClicks.summary.totalCount
    const totalCodeIntelHoverEvents =
        codeIntel.searchBasedEvents.summary.totalCount + codeIntel.preciseEvents.summary.totalCount

    const totalCodeIntelHoursSaved =
        (codeIntel.inAppEvents.summary.totalCount * 0.5 +
            codeIntel.codeHostEvents.summary.totalCount * 1.5 +
            Math.floor(
                (codeIntel.crossRepoEvents.summary.totalCount * totalCodeIntelEvents * 3) / totalCodeIntelHoverEvents ||
                    0
            ) +
            Math.floor(
                (codeIntel.preciseEvents.summary.totalCount * totalCodeIntelEvents) / totalCodeIntelHoverEvents || 0
            )) /
        60

    const totalBatchChangesEvents = batchChanges.changesetsMerged.summary.totalCount
    const totalBatchChangesHoursSaved = (totalBatchChangesEvents * 15) / 60

    const totalNotebooksEvents = notebooks.views.summary.totalCount
    const totalNotebooksHoursSaved = (totalNotebooksEvents * 5) / 60

    const totalExtensionsEvents =
        extensions.vscode.summary.totalCount +
        extensions.jetbrains.summary.totalCount +
        extensions.browser.summary.totalCount

    const totalExtensionsHoursSaved =
        (extensions.vscode.summary.totalCount * 3 +
            extensions.jetbrains.summary.totalCount * 1.5 +
            extensions.browser.summary.totalCount * 0.5) /
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
        if (dateRange === AnalyticsDateRange.LAST_WEEK) {
            return totalHoursSaved * 52
        }
        if (dateRange === AnalyticsDateRange.LAST_MONTH) {
            return totalHoursSaved * 12
        }
        if (dateRange === AnalyticsDateRange.LAST_THREE_MONTHS) {
            return (totalHoursSaved * 12) / 3
        }
        return totalHoursSaved
    })()

    const disableCodeSearchItems = !window.context?.codeSearchEnabledOnInstance

    return (
        <div>
            <H3 className="mb-3">Developer time saved</H3>
            <div className={classNames(styles.statsBox, 'p-4 mb-3')}>
                <div className="d-flex">
                    <ValueLegendItem
                        value={users.activity.summary.totalUniqueUsers}
                        className={classNames('flex-1', styles.borderRight)}
                        description="Active Users"
                        color="var(--body-color)"
                        tooltip="Currently registered users using the application in the selected timeframe."
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
                {showAnnualProjection && (
                    <div className="d-flex flex-column align-items-center mt-4">
                        <H2>
                            Annual projection:{' '}
                            <span className={styles.purple}>{formatNumber(projectedHoursSaved)} hours</span> saved*
                        </H2>
                        <Text as="span" className="text-muted">
                            * Based on{' '}
                            {dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
                                ? 'last 3 months'
                                : dateRange === AnalyticsDateRange.LAST_MONTH
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
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/search'}>
                                <Text as="span" className="d-flex align-items-center">
                                    <Icon svgPath={mdiMagnify} size="md" aria-label="Code Search" className="mr-1" />
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
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/code-intel'}>
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
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/batch-changes'}>
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
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/notebooks'}>
                                <Text as="span" className="d-flex align-items-center">
                                    <Icon svgPath={mdiBookOutline} size="md" aria-label="Notebooks" className="mr-1" />
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
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/extensions'}>
                                <Text as="span" className="d-flex align-items-center">
                                    <Icon
                                        svgPath={mdiPuzzleOutline}
                                        size="md"
                                        aria-label="Extensions"
                                        className="mr-1"
                                    />
                                    Search extensions
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
                    <tr>
                        <td className="text-left">
                            <Link to={disableCodeSearchItems ? '/search' : '/site-admin/analytics/code-insights'}>
                                <Text as="span" className="d-flex align-items-center">
                                    <Icon svgPath={mdiPoll} size="md" aria-label="Extensions" className="mr-1" />
                                    Code insights
                                </Text>
                            </Link>
                        </td>
                        <Tooltip content="Coming soon">
                            <Text className="cursor-pointer" as="td" weight="bold">
                                ...*
                            </Text>
                        </Tooltip>
                        <Tooltip content="Coming soon">
                            <Text className="cursor-pointer" as="td" weight="bold">
                                ...*
                            </Text>
                        </Tooltip>
                    </tr>
                </tbody>
            </table>
        </div>
    )
}
