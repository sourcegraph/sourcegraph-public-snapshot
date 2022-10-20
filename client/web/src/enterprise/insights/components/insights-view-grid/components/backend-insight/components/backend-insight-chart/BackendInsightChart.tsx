import { FC, MouseEvent, HTMLAttributes, useMemo, useRef, useState } from 'react'

import { mdiAlertCircle } from '@mdi/js'
import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import {
    BarChart,
    Icon,
    LegendItem,
    LegendList,
    Link,
    ScrollBox,
    Tooltip,
    Button,
    useOnClickOutside, useKeyboard
} from '@sourcegraph/wildcard'

import { UseSeriesToggleReturn } from '../../../../../../../../insights/utils/use-series-toggle'
import { SeriesBasedChartTypes, SeriesChart } from '../../../../../views'
import { BackendInsightData, BackendInsightSeries, InsightContentType } from '../../types'
import { BackendAlertOverlay } from '../backend-insight-alerts/BackendInsightAlerts'

import { getLineColor, hasNoData, hasSeriesError, isManyKeysInsight } from './selectors'

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

interface BackendInsightChartProps extends HTMLAttributes<HTMLDivElement> {
    data: BackendInsightData
    seriesToggleState: UseSeriesToggleReturn
    showSeriesErrors: boolean
    isInProgress: boolean
    isLocked: boolean
    isZeroYAxisMin: boolean
    onDatumClick: () => void
}

export const BackendInsightChart: FC<BackendInsightChartProps> = props => {
    const { data, seriesToggleState, showSeriesErrors, isInProgress, isLocked, isZeroYAxisMin, className, onDatumClick } = props

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
                            <SeriesLegends series={data.series} seriesToggleState={seriesToggleState} showSeriesErrors={showSeriesErrors}/>
                        </ScrollBox>
                    )}
                </>
            )}
        </div>
    )
}

interface SeriesLegendsProps {
    series: BackendInsightSeries[]
    seriesToggleState: UseSeriesToggleReturn
    showSeriesErrors: boolean
}

const SeriesLegends: FC<SeriesLegendsProps> = props => {
    const { series, seriesToggleState, showSeriesErrors } = props
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
                >
                    {item.name}
                    { showSeriesErrors && hasSeriesError(item) && <BackendInsightTimoutIcon/>}
                </LegendItem>
            ))}
        </LegendList>
    )
}

/**
 * Renders timeout icon and interactive tooltip with addition info about timeout
 * error. Note: It's exported because it's also used in the backend insight card.
 */
export const BackendInsightTimoutIcon: FC = () => {
    const [open, setOpen] = useState(false)
    const targetRef = useRef<HTMLButtonElement>(null)

    // We have to implement it locally because radix tooltip doesn't expose
    // enough API to support custom appearance logic (in this case click trigger logic)
    useKeyboard({ detectKeys: ['Escape'] }, () => setOpen(false))
    useOnClickOutside(targetRef,() => setOpen(false))

    const handleIconClick = (event: MouseEvent<HTMLButtonElement>): void => {
        // Catch event and prevent bubbling in order to prevent series toggle on/off
        // series action.
        event.stopPropagation()
        setOpen(!open)
    }

    return (
        <Tooltip
            open={open}
            content={
                <>
                    Calculating some points on this insight exceeded the timeout limit. Results may be incomplete.{' '}
                    <Link to="/" target="_blank" rel="noopener">
                        Troubleshoot
                    </Link>
                </>
            }
        >
            <Button ref={targetRef} variant='icon' className={styles.timeoutIcon} onClick={handleIconClick}>
                <Icon
                    aria-label="Insight is timeout"
                    svgPath={mdiAlertCircle}
                    color='var(--icon-color)'
                />
            </Button>
        </Tooltip>
    )
}
