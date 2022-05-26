import React, { Dispatch, SetStateAction } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { useDebounce } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList } from '../../../../../../../../charts'
import { ScrollBox } from '../../../../../../../../views/components/view/content/chart-view-content/charts/line/components/scroll-box/ScrollBox'
import { BackendInsightData } from '../../../../../../core'
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
    className?: string
    onDatumClick: () => void
    toggle: (id: string) => void
    isSeriesSelected: (id: string) => boolean
    isSeriesHovered: (id: string) => boolean
    setHoveredId: Dispatch<SetStateAction<string | undefined>>
}

export function BackendInsightChart<Datum>(props: BackendInsightChartProps<Datum>): React.ReactElement {
    const {
        locked,
        isFetchingHistoricalData,
        content,
        className,
        onDatumClick,
        toggle,
        isSeriesSelected,
        isSeriesHovered,
        setHoveredId,
    } = props
    const { ref, width = 0 } = useDebounce(useResizeObserver(), 100)

    const hasViewManySeries = content.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
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
                                    hasNoData={content.series.every(series => series.data.length === 0)}
                                    isFetchingHistoricalData={isFetchingHistoricalData}
                                    className={styles.alertOverlay}
                                />

                                <SeriesChart
                                    type={SeriesBasedChartTypes.Line}
                                    width={parent.width}
                                    height={parent.height}
                                    locked={locked}
                                    className={styles.chart}
                                    onDatumClick={onDatumClick}
                                    isSeriesSelected={isSeriesSelected}
                                    isSeriesHovered={isSeriesHovered}
                                    {...content}
                                />
                            </>
                        )}
                    </ParentSize>

                    <ScrollBox className={styles.legendListContainer} onMouseLeave={() => setHoveredId(undefined)}>
                        <LegendList className={styles.legendList}>
                            {content.series.map(series => (
                                <LegendItem
                                    key={series.id as string}
                                    color={getLineColor(series)}
                                    name={series.name}
                                    selected={isSeriesSelected(`${series.id}`)}
                                    hovered={isSeriesHovered(`${series.id}`)}
                                    className={styles.legendListItem}
                                    onClick={() => toggle(`${series.id}`)}
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
