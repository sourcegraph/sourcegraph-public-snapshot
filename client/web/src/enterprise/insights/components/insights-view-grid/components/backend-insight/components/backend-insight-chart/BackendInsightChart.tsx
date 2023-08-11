import React, { type FC, type MouseEvent, useMemo } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { Button, BarChart, LegendItem, LegendList, LegendItemPoint, ScrollBox } from '@sourcegraph/wildcard'

import type { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import type { BackendInsightData, BackendInsightSeries, InsightContent } from '../../../../../../core'
import { InsightContentType } from '../../../../../../core/types/insight/common'
import { SeriesBasedChartTypes, SeriesChart } from '../../../../../views'
import { BackendAlertOverlay, InsightSeriesIncompleteAlert } from '../backend-insight-alerts/BackendInsightAlerts'

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
export const MINIMAL_HORIZONTAL_LAYOUT_WIDTH = 460

/**
 * Even if you have a big enough width for putting legend aside (see {@link MINIMAL_HORIZONTAL_LAYOUT_WIDTH})
 * we should enable this mode only if line chart has more than N series
 */
export const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3

interface BackendInsightChartProps<Datum> extends BackendInsightData {
    locked: boolean
    zeroYAxisMin: boolean
    seriesToggleState: UseSeriesToggleReturn
    className?: string
    onDatumClick: () => void
}

export function BackendInsightChart<Datum>(props: BackendInsightChartProps<Datum>): React.ReactElement {
    const { data, isFetchingHistoricalData, locked, zeroYAxisMin, seriesToggleState, className, onDatumClick } = props

    const { ref, width = 0 } = useResizeObserver()

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
            {width > 0 && (
                <>
                    <ParentSize
                        debounceTime={0}
                        enableDebounceLeadingCall={true}
                        className={styles.responsiveContainer}
                    >
                        {parent =>
                            // Render chart element only when we have real non-empty parent sizes
                            // otherwise, the first chart render happens on a not fully rendered
                            // element that causes the element's flickering
                            parent.height * parent.width !== 0 && (
                                <>
                                    <BackendAlertOverlay
                                        hasNoData={isEmptyDataset}
                                        isFetchingHistoricalData={isFetchingHistoricalData}
                                        className={styles.alertOverlay}
                                    />

                                    {data.type === InsightContentType.Series ? (
                                        <SeriesChart
                                            type={SeriesBasedChartTypes.Line}
                                            width={parent.width}
                                            height={parent.height}
                                            locked={locked}
                                            className={styles.chart}
                                            onDatumClick={onDatumClick}
                                            zeroYAxisMin={zeroYAxisMin}
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
                            )
                        }
                    </ParentSize>

                    {isSeriesLikeInsight && (
                        <ScrollBox className={styles.legendListContainer}>
                            <SeriesLegends series={data.series} seriesToggleState={seriesToggleState} />
                        </ScrollBox>
                    )}
                </>
            )}
        </div>
    )
}

const isManyKeysInsight = (data: InsightContent<any>): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    }

    return data.content.data.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
}

const hasNoData = (data: InsightContent<any>): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.series.every(series => series.data.length === 0)
    }

    // If all datum have zero matches render no data layout. We need to
    // handle it explicitly on the frontend since backend returns manually
    // defined series with empty points in case of no matches for generated
    // series.
    return data.content.data.every(datum => datum.value === 0)
}

interface SeriesLegendsProps {
    series: BackendInsightSeries<any>[]
    seriesToggleState: UseSeriesToggleReturn
}

const SeriesLegends: FC<SeriesLegendsProps> = props => {
    const { series, seriesToggleState } = props

    // Non-interactive static legend list
    if (series.length <= 1) {
        return (
            <LegendList className={styles.legendList}>
                {series.map(item => (
                    <LegendItem key={item.id as string} color={item.color}>
                        <LegendItemPoint color={item.color} />
                        {item.name}
                        {item.alerts.length > 0 && <InsightSeriesIncompleteAlert series={item} />}
                    </LegendItem>
                ))}
            </LegendList>
        )
    }

    const { setHoveredId, isSeriesSelected, isSeriesHovered, toggle } = seriesToggleState

    // Interactive legends list
    return (
        <LegendList
            className={styles.legendList}
            // Prevent accidental dragging events
            onMouseDown={(event: MouseEvent<HTMLElement>) => event.stopPropagation()}
        >
            {series.map(item => (
                <LegendItem
                    key={item.id as string}
                    active={isSeriesHovered(`${item.id}`) || isSeriesSelected(`${item.id}`)}
                >
                    <Button
                        role="checkbox"
                        aria-checked={isSeriesSelected(`${item.id}`)}
                        className={classNames(styles.legendListItem, styles.legendListItemInteractive)}
                        onPointerEnter={() => setHoveredId(`${item.id}`)}
                        onPointerLeave={() => setHoveredId(undefined)}
                        onFocus={() => setHoveredId(`${item.id}`)}
                        onBlur={() => setHoveredId(undefined)}
                        onClick={() =>
                            toggle(
                                `${item.id}`,
                                series.map(series => `${series.id}`)
                            )
                        }
                    >
                        <LegendItemPoint color={item.color} />
                        {item.name}
                    </Button>
                    {item.alerts.length > 0 && (
                        <InsightSeriesIncompleteAlert series={item} className={styles.legendIncompleteAlert} />
                    )}
                </LegendItem>
            ))}
        </LegendList>
    )
}
