import { type FC, useEffect, useState, memo } from 'react'

import { mdiArrowExpand } from '@mdi/js'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon } from '@sourcegraph/wildcard'

import type { SearchAggregationMode, SearchPatternType } from '../../../graphql-operations'
import {
    AggregationChartCard,
    AggregationModeControls,
    AggregationLimitLabel,
    AggregationUIMode,
    useAggregationSearchMode,
    useAggregationUIMode,
    useSearchAggregationData,
    isNonExhaustiveAggregationResults,
    GroupResultsPing,
} from '../components/aggregation'

import styles from './SearchAggregations.module.scss'

interface SearchAggregationsProps extends TelemetryProps, TelemetryV2Props {
    /**
     * Current submitted query, note that this query isn't a live query
     * that is synced with typed query in the search box, this query is submitted
     * see `searchQueryFromURL` state in the global query Zustand store.
     */
    query: string

    /** Current search query pattern type. */
    patternType: SearchPatternType

    /** Whether to proactively load and display search aggregations */
    proactive: boolean

    caseSensitive: boolean

    /**
     * Emits whenever a user clicks one of aggregation chart segments (bars).
     * That should update the query and re-trigger search (but this should be connected
     * to this UI through its consumer)
     */
    onQuerySubmit: (newQuery: string, updatedSearchQuery: string) => void
}

export const SearchAggregations: FC<SearchAggregationsProps> = memo(props => {
    const { query, patternType, proactive, caseSensitive, telemetryService, telemetryRecorder, onQuerySubmit } = props

    const [extendedTimeout, setExtendedTimeoutLocal] = useState(false)

    const [, setAggregationUIMode] = useAggregationUIMode()
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const { data, error, loading } = useSearchAggregationData({
        query,
        patternType,
        aggregationMode,
        proactive,
        caseSensitive,
        extendedTimeout,
        telemetryService,
        telemetryRecorder,
    })

    // When query is updated reset extendedTimeout as per business rules
    useEffect(() => setExtendedTimeoutLocal(false), [query])

    const handleExtendTimeout = (): void => setExtendedTimeoutLocal(true)

    const handleBarLinkClick = (query: string, index: number): void => {
        // Clearing the aggregation mode on drill down would provide a better experience
        // in most cases and preserve the desired behavior of the capture group search
        // when the original query had multiple capture groups
        const updatedSearchQuery = setAggregationMode(null)

        onQuerySubmit(query, updatedSearchQuery)
        telemetryService.log(
            GroupResultsPing.ChartBarClick,
            { aggregationMode, index, uiMode: 'sidebar' },
            { aggregationMode, index, uiMode: 'sidebar' }
        )
        telemetryRecorder.recordEvent(GroupResultsPing.ChartBarClick, 'clicked', {
            privateMetadata: {
                aggregationMode,
                index,
                uiMode: 'sidebar',
            },
        })
    }

    const handleBarHover = (): void => {
        telemetryService.log(
            GroupResultsPing.ChartBarHover,
            { aggregationMode, uiMode: 'sidebar' },
            { aggregationMode, uiMode: 'sidebar' }
        )
        telemetryRecorder.recordEvent(GroupResultsPing.ChartBarHover, 'hovered', {
            privateMetadata: {
                aggregationMode,
                uiMode: 'sidebar',
            },
        })
    }

    const handleExpandClick = (): void => {
        setAggregationUIMode(AggregationUIMode.SearchPage)
        telemetryService.log(GroupResultsPing.ExpandFullViewPanel, { aggregationMode }, { aggregationMode })
        telemetryRecorder.recordEvent(GroupResultsPing.ExpandFullViewPanel, 'clicked', {
            privateMetadata: {
                aggregationMode,
            },
        })
    }

    const handleAggregationModeChange = (mode: SearchAggregationMode): void => {
        setAggregationMode(mode)
        telemetryService.log(
            GroupResultsPing.ModeClick,
            { aggregationMode: mode, uiMode: 'sidebar' },
            { aggregationMode: mode, uiMode: 'sidebar' }
        )
        telemetryRecorder.recordEvent(GroupResultsPing.ModeClick, 'clicked', {
            privateMetadata: {
                aggregationMode: mode,
                uiMode: 'sidebar',
            },
        })
    }

    const handleAggregationModeHover = (aggregationMode: SearchAggregationMode, available: boolean): void => {
        if (!available) {
            telemetryService.log(
                GroupResultsPing.ModeDisabledHover,
                { aggregationMode, uiMode: 'sidebar' },
                { aggregationMode, uiMode: 'sidebar' }
            )
            telemetryRecorder.recordEvent(GroupResultsPing.ModeDisabledHover, 'hovered', {
                privateMetadata: {
                    aggregationMode,
                    uiMode: 'sidebar',
                },
            })
        }
    }

    return (
        <div className="pt-2">
            <AggregationModeControls
                availability={data?.searchQueryAggregate?.modeAvailability}
                loading={loading}
                mode={aggregationMode}
                size="sm"
                onModeChange={handleAggregationModeChange}
                onModeHover={handleAggregationModeHover}
            />

            {(proactive || aggregationMode !== null) && (
                <>
                    <AggregationChartCard
                        aria-label="Sidebar search aggregation chart"
                        data={data?.searchQueryAggregate?.aggregations}
                        loading={loading}
                        error={error}
                        mode={aggregationMode}
                        showLoading={extendedTimeout}
                        className={styles.chartContainer}
                        onBarLinkClick={handleBarLinkClick}
                        onBarHover={handleBarHover}
                        onExtendTimeout={handleExtendTimeout}
                    />

                    <footer className={styles.actions}>
                        <Button
                            variant="secondary"
                            size="sm"
                            outline={true}
                            className={styles.detailsAction}
                            data-testid="expand-aggregation-ui"
                            onClick={handleExpandClick}
                        >
                            <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Expand
                        </Button>

                        {isNonExhaustiveAggregationResults(data) && <AggregationLimitLabel size="sm" />}
                    </footer>
                </>
            )}
        </div>
    )
})

SearchAggregations.displayName = 'SearchAggregations'
