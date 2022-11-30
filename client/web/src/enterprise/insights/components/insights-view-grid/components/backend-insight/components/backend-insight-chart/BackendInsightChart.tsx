import React, { FC, MouseEvent, useMemo } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { BarChart, ScrollBox, LegendList, LegendItem, Button, Series, LegendItemPoint } from '@sourcegraph/wildcard'

import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { BackendInsightData, InsightContent } from '../../../../../../core'
import { InsightContentType } from '../../../../../../core/types/insight/common'
import { SeriesBasedChartTypes, SeriesChart } from '../../../../../views'
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
export const MINIMAL_HORIZONTAL_LAYOUT_WIDTH = 460

/**
 * Even if you have a big enough width for putting legend aside (see {@link MINIMAL_HORIZONTAL_LAYOUT_WIDTH})
 * we should enable this mode only if line chart has more than N series
 */
export const MINIMAL_SERIES_FOR_ASIDE_LEGEND = 3

interface BackendInsightChartProps<Datum> extends BackendInsightData {
    locked: boolean
    zeroYAxisMin: boolean
    className?: string
    onDatumClick: () => void
    seriesToggleState: UseSeriesToggleReturn
}

export function BackendInsightChart<Datum>(props: BackendInsightChartProps<Datum>): React.ReactElement {
    const { locked, isFetchingHistoricalData, data, zeroYAxisMin, className, onDatumClick, seriesToggleState } = props
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
                                        {...data.content}
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
                        <ScrollBox className={styles.legendListContainer}>
                            <SeriesLegends series={data.content.series} seriesToggleState={seriesToggleState} />
                        </ScrollBox>
                    )}
                </>
            )}
        </div>
    )
}

const isManyKeysInsight = (data: InsightContent<any>): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.content.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    }

    return data.content.data.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
}

const hasNoData = (data: InsightContent<any>): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.content.series.every(series => series.data.length === 0)
    }

    // If all datum have zero matches render no data layout. We need to
    // handle it explicitly on the frontend since backend returns manually
    // defined series with empty points in case of no matches for generated
    // series.
    return data.content.data.every(datum => datum.value === 0)
}

interface SeriesLegendsProps {
    series: Series<any>[]
    seriesToggleState: UseSeriesToggleReturn
}

const SeriesLegends: FC<SeriesLegendsProps> = props => {
    const { series, seriesToggleState } = props

    // Non-interactive static legend list
    if (series.length <= 1) {
        return (
            <LegendList className={styles.legendList}>
                {series.map(item => (
                    <LegendItem key={item.id as string} name={item.name} color={item.color} />
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
                        className={styles.legendListItem}
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
                </LegendItem>
            ))}
        </LegendList>
    )
}
