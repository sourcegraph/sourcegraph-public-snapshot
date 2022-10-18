import { Suspense, HTMLAttributes, ReactElement, MouseEvent } from 'react'

import { mdiPlay } from '@mdi/js'

import { ErrorAlert, ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { pluralize } from '@sourcegraph/common'
import { NotAvailableReasonType, SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Text, Link, Tooltip, Button, Icon } from '@sourcegraph/wildcard'

import { SearchAggregationDatum, GetSearchAggregationResult } from '../../../../../../graphql-operations'

import { AggregationChartProps, AggregationTextContent, AggregationContent } from './components'

import styles from './AggregationChartCard.module.scss'

const LazyAggregationChart = lazyComponent<AggregationChartProps<SearchAggregationDatum>, string>(
    () => import('./components/AggregationChart'),
    'AggregationChart'
)
LazyAggregationChart.displayName = 'LazyAggregationChart'

/** Set custom value for minimal rotation angle for X ticks in sidebar UI panel mode. */
const MIN_X_TICK_ROTATION = 30
const MAX_SHORT_LABEL_WIDTH = 8
const MAX_LABEL_WIDTH = 16
const MAX_BARS_FULL_MODE = 30
const MAX_BARS_PREVIEW_MOD = 10

const getName = (datum: SearchAggregationDatum): string => datum.label ?? ''
const getValue = (datum: SearchAggregationDatum): number => datum.count
const getLink = (datum: SearchAggregationDatum): string => datum.query ?? ''
const getColor = (): string => 'var(--primary)'

/**
 * Nested aggregation results types from {@link AGGREGATION_SEARCH_QUERY} GQL query
 */
type SearchAggregationResult = GetSearchAggregationResult['searchQueryAggregate']['aggregations']

function getAggregationError(
    aggregation?: SearchAggregationResult
): { error: Error; type: NotAvailableReasonType } | undefined {
    if (aggregation?.__typename === 'SearchAggregationNotAvailable') {
        return { error: new Error(aggregation.reason), type: aggregation.reasonType }
    }

    return
}

export function getAggregationData(aggregations: SearchAggregationResult, limit?: number): SearchAggregationDatum[] {
    switch (aggregations?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
        case 'NonExhaustiveSearchAggregationResult':
            return limit !== undefined ? aggregations.groups.slice(0, limit) : aggregations.groups

        default:
            return []
    }
}

export function getOtherGroupCount(aggregations: SearchAggregationResult, limit: number): number {
    switch (aggregations?.__typename) {
        case 'ExhaustiveSearchAggregationResult':
            return (aggregations.otherGroupCount ?? 0) + Math.max(aggregations.groups.length - limit, 0)
        case 'NonExhaustiveSearchAggregationResult':
            return (aggregations.approximateOtherGroupCount ?? 0) + Math.max(aggregations.groups.length - limit, 0)

        default:
            return 0
    }
}

interface AggregationChartCardProps extends HTMLAttributes<HTMLDivElement> {
    data?: SearchAggregationResult
    error?: Error
    loading: boolean
    mode?: SearchAggregationMode | null
    size?: 'sm' | 'md'
    showLoading?: boolean
    onBarLinkClick?: (query: string, barIndex: number) => void
    onBarHover?: () => void
    onExtendTimeout: () => void
}

export function AggregationChartCard(props: AggregationChartCardProps): ReactElement | null {
    const {
        data,
        error,
        loading,
        mode,
        className,
        size = 'sm',
        showLoading = false,
        'aria-label': ariaLabel,
        onBarLinkClick,
        onBarHover,
        onExtendTimeout,
    } = props

    if (loading) {
        return (
            <AggregationTextContent size={size} className={className}>
                {showLoading && <span className={styles.loading}>Loading</span>}
            </AggregationTextContent>
        )
    }

    // Internal error
    if (error) {
        return (
            <AggregationContent size={size} className={className}>
                <ErrorAlert error={error} className={styles.errorAlert} />
            </AggregationContent>
        )
    }

    const aggregationError = getAggregationError(data)

    if (aggregationError) {
        return (
            <AggregationTextContent size={size} className={className}>
                {aggregationError.type === NotAvailableReasonType.TIMEOUT_EXTENSION_AVAILABLE ? (
                    <Button variant="link" size="sm" className={styles.errorButton} onClick={onExtendTimeout}>
                        <Icon aria-hidden={true} svgPath={mdiPlay} className="mr-1" />
                        Run aggregation
                    </Button>
                ) : (
                    <>
                        We couldn’t provide an aggregation for this query.{' '}
                        <ErrorMessage error={aggregationError.error} />{' '}
                        <Link to="/help/code_insights/explanations/search_results_aggregations">Learn more</Link>
                    </>
                )}
            </AggregationTextContent>
        )
    }

    if (!data) {
        return null
    }

    const maxBarsLimit = size === 'sm' ? MAX_BARS_PREVIEW_MOD : MAX_BARS_FULL_MODE
    const aggregationData = getAggregationData(data, maxBarsLimit)
    const missingCount = getOtherGroupCount(data, maxBarsLimit)

    if (aggregationData.length === 0) {
        return (
            <AggregationTextContent size={size} className={className}>
                No data to display
            </AggregationTextContent>
        )
    }

    const handleDatumLinkClick = (event: MouseEvent, datum: SearchAggregationDatum, index: number): void => {
        event.preventDefault()
        onBarLinkClick?.(getLink(datum), index)
    }

    return (
        <AggregationContent className={className}>
            <Suspense>
                <LazyAggregationChart
                    aria-label={ariaLabel}
                    data={aggregationData}
                    mode={mode}
                    minAngleXTick={size === 'md' ? 0 : MIN_X_TICK_ROTATION}
                    maxXLabelLength={size === 'md' ? MAX_LABEL_WIDTH : MAX_SHORT_LABEL_WIDTH}
                    getDatumValue={getValue}
                    getDatumColor={getColor}
                    getDatumName={getName}
                    getDatumLink={getLink}
                    onDatumLinkClick={handleDatumLinkClick}
                    onDatumHover={onBarHover}
                    className={styles.chart}
                />

                {!!missingCount && (
                    <Tooltip
                        content={`There ${pluralize('is', missingCount, 'are')} ${missingCount} more ${pluralize(
                            'group',
                            missingCount,
                            'groups'
                        )} not shown.`}
                    >
                        <Text size="small" className={styles.missingLabelCount}>
                            +{missingCount}
                        </Text>
                    </Tooltip>
                )}
            </Suspense>
        </AggregationContent>
    )
}
