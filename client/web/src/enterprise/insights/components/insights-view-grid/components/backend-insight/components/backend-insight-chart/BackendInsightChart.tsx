import { FC, HTMLAttributes, useMemo } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { BarChart, ScrollBox, LegendList, LegendItem, Series } from '@sourcegraph/wildcard'

import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { SeriesBasedChartTypes, SeriesChart } from '../../../../../views'
import { BackendInsightData, InsightContentType } from '../../types'
import { BackendAlertOverlay } from '../backend-insight-alerts/BackendInsightAlerts'

import styles from './BackendInsightChart.module.scss'

/**
 * If width of the chart is less than this var width value we should put the legend
 * block below the chart block
 *
 * ```
 * Less than 450px - put legend below      Chart block has enough space - render legend aside
 * ▲                                       ▲
 * │             ● ●                       │           ●    ● Item 1
 * │      ●     ●                          │    ●     ●     ● Item 2
 * │     ●  ●  ●                           │   ●  ●  ●
 * │    ●    ●                             │  ●    ●
 * │   ●                                   │ ●
 * │                                       │
 * └─────────────────▶                     └─────────────▶
 * ● Item 1 ● Item 2
 * ```
 */
const MINIMAL_HORIZONTAL_LAYOUT_WIDTH = 460

/**
 * Even if you have a big enough width for putting legend aside (see {@link MINIMAL_HORIZONTAL_LAYOUT_WIDTH})
 * we should enable this mode only if line chart has more than N series
 */
const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3

interface BackendInsightChartProps extends HTMLAttributes<HTMLDivElement> {
    data: BackendInsightData
    seriesToggleState: UseSeriesToggleReturn
    isInProgress: boolean
    isLocked: boolean
    isZeroYAxisMin: boolean
    onDatumClick: () => void
}

export const BackendInsightChart: FC<BackendInsightChartProps> = props => {
    const { data, seriesToggleState, isInProgress, isLocked, isZeroYAxisMin, className, onDatumClick } = props

    const { ref, width = 0 } = useResizeObserver()
    const { setHoveredId } = seriesToggleState

    const isEmptyDataset = useMemo(() => hasNoData(data), [data])

    const hasViewManySeries = isManyKeysInsight(data)
    const hasEnoughXSpace = width >= MINIMAL_HORIZONTAL_LAYOUT_WIDTH
    const isHorizontalMode = hasViewManySeries && hasEnoughXSpace
    const isSeriesLikeInsight = data.type === InsightContentType.Series

    return (
        <div
            ref={ref}
            className={classNames(className, styles.root, {
                [styles.rootHorizontal]: isHorizontalMode,
                [styles.rootWithLegend]: isSeriesLikeInsight,
            })}
        >
            {width && (
                <>
                    <ParentSize
                        debounceTime={0}
                        enableDebounceLeadingCall={true}
                        className={styles.responsiveContainer}
                    >
                        {parent => (
                            <>
                                <BackendAlertOverlay
                                    hasNoData={isEmptyDataset}
                                    isFetchingHistoricalData={isInProgress}
                                    className={styles.alertOverlay}
                                />

                                {data.type === InsightContentType.Series ? (
                                    <SeriesChart
                                        type={SeriesBasedChartTypes.Line}
                                        width={parent.width}
                                        height={parent.height}
                                        locked={isLocked}
                                        className={styles.chart}
                                        onDatumClick={onDatumClick}
                                        zeroYAxisMin={isZeroYAxisMin}
                                        seriesToggleState={seriesToggleState}
                                        series={data.series}
                                    />
                                ) : (
                                    <BarChart
                                        aria-label="Bar chart"
                                        width={parent.width}
                                        height={parent.height}
                                        {...data.content}
                                    />
                                )}
                            </>
                        )}
                    </ParentSize>

                    {isSeriesLikeInsight && (
                        <ScrollBox className={styles.legendListContainer} onMouseLeave={() => setHoveredId(undefined)}>
                            <SeriesLegends series={data.series} seriesToggleState={seriesToggleState} />
                        </ScrollBox>
                    )}
                </>
            )}
        </div>
    )
}

const isManyKeysInsight = (data: BackendInsightData): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    }

    return data.content.data.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
}

const hasNoData = (data: BackendInsightData): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.every(series => series.data.length === 0)
    }

    // If all datum have zero matches render no data layout. We need to
    // handle it explicitly on the frontend since backend returns manually
    // defined series with empty points in case of no matches for generated
    // series.
    return data.content.data.every(datum => datum.value === 0)
}

function getLineColor(series: Series<any>): string {
    return series.color ?? 'var(--gray-07)'
}

interface SeriesLegendsProps {
    series: Series<any>[]
    seriesToggleState: UseSeriesToggleReturn
}

const SeriesLegends: FC<SeriesLegendsProps> = props => {
    const { series, seriesToggleState } = props

    const { setHoveredId, isSeriesSelected, isSeriesHovered, toggle } = seriesToggleState

    return (
        <LegendList className={styles.legendList}>
            {series.map(item => (
                <LegendItem
                    key={item.id as string}
                    color={getLineColor(item)}
                    name={item.name}
                    selected={isSeriesSelected(`${item.id}`)}
                    hovered={isSeriesHovered(`${item.id}`)}
                    className={classNames(styles.legendListItem, {
                        [styles.clickable]: series.length > 1,
                    })}
                    onClick={() =>
                        toggle(
                            `${item.id}`,
                            series.map(series => `${series.id}`)
                        )
                    }
                    onMouseEnter={() => setHoveredId(`${item.id}`)}
                    // prevent accidental dragging events
                    onMouseDown={event => event.stopPropagation()}
                />
            ))}
        </LegendList>
    )
}
