import React from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { useDebounce } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ScrollBox, Series } from '../../../../../../../../charts'
import { BarChart } from '../../../../../../../../charts/components/bar-chart/BarChart'
import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { ComputeInsightData } from '../../../../../../core'
import { BackendAlertOverlay } from '../backend-insight-alerts/BackendInsightAlerts'

import styles from './ComputeInsightChart.module.scss'

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

interface ComputeInsightChartProps<Datum> extends ComputeInsightData {
    className?: string
    seriesToggleState: UseSeriesToggleReturn
}

export function ComputeInsightChart<Datum>(props: ComputeInsightChartProps<Datum>): React.ReactElement {
    const { isFetchingHistoricalData, content, className, seriesToggleState } = props
    const { ref, width = 0 } = useDebounce(useResizeObserver(), 100)
    const { setHoveredId, isSeriesSelected, isSeriesHovered } = seriesToggleState

    const hasViewManySeries = content.data.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
    const hasEnoughXSpace = width >= MINIMAL_HORIZONTAL_LAYOUT_WIDTH

    const isHorizontalMode = hasViewManySeries && hasEnoughXSpace

    return (
        <div ref={ref} className={classNames(className, styles.root, { [styles.rootHorizontal]: isHorizontalMode })}>
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
                                    hasNoData={content.data.length === 0}
                                    isFetchingHistoricalData={isFetchingHistoricalData}
                                    className={styles.alertOverlay}
                                />

                                <BarChart
                                    width={parent.width}
                                    height={parent.height}
                                    getDatumColor={content.getDatumColor}
                                    getDatumName={content.getDatumName}
                                    getDatumValue={content.getDatumValue}
                                    data={content.data}
                                />
                            </>
                        )}
                    </ParentSize>

                    <ScrollBox className={styles.legendListContainer} onMouseLeave={() => setHoveredId(undefined)}>
                        <LegendList className={styles.legendList}>
                            {content.data.map(series => (
                                <LegendItem
                                    key={series.id as string}
                                    color={getLineColor(series)}
                                    name={series.name}
                                    selected={isSeriesSelected(`${series.id}`)}
                                    hovered={isSeriesHovered(`${series.id}`)}
                                    className={classNames(styles.legendListItem, {
                                        [styles.clickable]: content.data.length > 1,
                                    })}
                                    onClick={() => seriesToggleState.toggle(`${series.id}`, mapSeriesIds(content.data))}
                                    onMouseEnter={() => setHoveredId(`${series.id}`)}
                                    // prevent accidental dragging events
                                    onMouseDown={event => event.stopPropagation()}
                                />
                            ))}
                        </LegendList>
                    </ScrollBox>
                </>
            )}
        </div>
    )
}

const mapSeriesIds = <D,>(series: Series<D>[]): string[] => series.map(series => `${series.id}`)
