import React, { FC, MouseEvent, useCallback, useMemo, useState } from 'react'

import { mdiAlertCircle } from '@mdi/js'
import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import {
    Link,
    Button,
    Icon,
    BarChart,
    LegendItem,
    LegendList,
    LegendItemPoint,
    ScrollBox,
    Tooltip,
    TooltipOpenEvent,
    TooltipOpenChangeReason,
} from '@sourcegraph/wildcard'

import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { BackendInsightData, BackendInsightSeries, InsightContent } from '../../../../../../core'
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
                                        series={data.content.series}
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
                        {item.errored && <BackendInsightTimeoutIcon />}
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
                    {item.errored && <BackendInsightTimeoutIcon />}
                </LegendItem>
            ))}
        </LegendList>
    )
}

interface BackendInsightTimeoutIconProps {
    timeoutLevel?: 'series' | 'insight'
}

/**
 * Renders timeout icon and interactive tooltip with addition info about timeout
 * error. Note: It's exported because it's also used in the backend insight card.
 */
export const BackendInsightTimeoutIcon: FC<BackendInsightTimeoutIconProps> = props => {
    const { timeoutLevel = 'series' } = props
    const [open, setOpen] = useState(false)

    const handleIconClick = (event: MouseEvent<HTMLButtonElement>): void => {
        // Catch event and prevent bubbling in order to prevent series toggle on/off
        // series action.
        event.stopPropagation()
        setOpen(!open)
    }

    const handleOpenChange = useCallback((event: TooltipOpenEvent): void => {
        switch (event.reason) {
            case TooltipOpenChangeReason.Esc:
            case TooltipOpenChangeReason.ClickOutside: {
                setOpen(event.isOpen)
            }
        }
    }, [])

    return (
        <Tooltip
            open={open}
            content={
                <>
                    {timeoutLevel === 'series'
                        ? 'Some points of this data series exceeded the time limit. Results may be incomplete.'
                        : 'Calculating some points on this insight exceeded the timeout limit. Results may be incomplete.'}{' '}
                    <Link
                        to="/help/code_insights/how-tos/Troubleshooting"
                        target="_blank"
                        rel="noopener"
                        className={styles.troubleshootLink}
                    >
                        Troubleshoot
                    </Link>
                </>
            }
            onOpenChange={handleOpenChange}
        >
            <Button variant="icon" className={styles.timeoutIcon} onClick={handleIconClick}>
                <Icon aria-label="Insight is timeout" svgPath={mdiAlertCircle} color="var(--icon-color)" />
            </Button>
        </Tooltip>
    )
}
