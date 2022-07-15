import React, { FC } from 'react'

import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { useDebounce } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ScrollBox, Series } from '../../../../../../../../charts'
import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { BackendInsightData, CategoricalChartContent, InsightContent } from '../../../../../../core'
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
    const { ref, width = 0 } = useDebounce(useResizeObserver(), 100)
    const { setHoveredId } = seriesToggleState

    const hasViewManySeries = isManyKeysInsight(data)
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
                                    hasNoData={hasData(data)}
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
                                ) : null}
                            </>
                        )}
                    </ParentSize>

                    <ScrollBox className={styles.legendListContainer} onMouseLeave={() => setHoveredId(undefined)}>
                        {data.type === InsightContentType.Series ? (
                            <SeriesLegends series={data.content.series} seriesToggleState={seriesToggleState} />
                        ) : (
                            <CategoricalLegends data={data.content} />
                        )}
                    </ScrollBox>
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

const hasData = (data: InsightContent<any>): boolean => {
    if (data.type === InsightContentType.Series) {
        return data.content.series.every(series => series.data.length === 0)
    }

    return data.content.data.length > 0
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

interface CategoricalLegendsProps {
    data: CategoricalChartContent<any>
}

const CategoricalLegends: FC<CategoricalLegendsProps> = props => {
    const { data } = props

    return (
        <LegendList className={styles.legendList}>
            {data.data.map(item => (
                <LegendItem
                    key={item.id as string}
                    color={data.getDatumColor(item) ?? 'gray'}
                    name={data.getDatumName(item)}
                    className={styles.legendListItem}
                    // prevent accidental dragging events
                    onMouseDown={event => event.stopPropagation()}
                />
            ))}
        </LegendList>
    )
}
