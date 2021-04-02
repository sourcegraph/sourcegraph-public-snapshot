import React, { ReactElement, useCallback, useMemo, useState, MouseEvent } from 'react';
import classnames from 'classnames';
import { LineChartContent } from 'sourcegraph';
import { useDebouncedCallback } from 'use-debounce';
import { curveLinear } from '@visx/curve';
import { ParentSize } from '@visx/responsive';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';
import {
    Axis,
    GlyphSeries,
    lightTheme,
    LineSeries,
    Tooltip,
    XYChart,
} from '@visx/xychart';

import { format } from 'd3-format';
import { timeFormat } from 'd3-time-format'
import { scaleLinear, scaleTime } from '@visx/scale';
import { GridColumns, GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import { GlyphDot } from '@visx/glyph'
import { EventHandlerParams } from '@visx/xychart/lib/types';

import { generateAccessors } from './helpers/generate-accessors'
import { getRangeWithPadding } from './helpers/get-range-with-padding'
import { getMinAndMax } from './helpers/get-min-max'
import { GlyphDotComponent } from './components/GlyphDot'
import { TooltipContent } from './components/TooltipContent'
import { onDatumClick } from '../types';
import isValidNumber from '@visx/xychart/lib/typeguards/isValidNumber';

// Chart configuration
const WIDTH_PER_TICK = 70;
const HEIGHT_PER_TICK = 40;
const MARGIN = { top: 10, left: 30, bottom: 20, right: 20 };

// Formatters
const dateFormatter = timeFormat('%d %b');
const formatDate = (date: Date): string => dateFormatter(date);

export interface LineChartProps extends Omit<LineChartContent<any, string>, 'chart'> {
    width: number;
    height: number;
    onDatumClick: onDatumClick;
}

function LineChartContentComponent(props: LineChartProps): ReactElement {
    const { width, height, data, series, xAxis, onDatumClick } = props;

    // derived
    const innerWidth = width - MARGIN.left - MARGIN.right;
    const innerHeight = height - MARGIN.top - MARGIN.bottom;

    const numberOfTicks = Math.max(1, Math.floor(innerWidth / WIDTH_PER_TICK))
    const numberOfTicksY = Math.max(1, Math.floor(innerHeight / HEIGHT_PER_TICK))

    const sortedData = useMemo(
        () => data.sort(
            (firstDatum, secondDatum) => firstDatum[xAxis.dataKey] - secondDatum[xAxis.dataKey]
        ),
        [data, xAxis]
    );
    const accessors = useMemo(
        () => generateAccessors(xAxis, series),
        [xAxis, series]
    );
    const scalesConfig = useMemo(
        () => {
            const scale = scaleLinear({
                domain: getRangeWithPadding(getMinAndMax(sortedData, accessors), 0.3),
                nice: true,
                zero: false,
            })

            const ticks = scale.ticks(numberOfTicksY);
            const firstTickValue = scale.ticks()[0];
            const lastTickValue = ticks[ticks.length - 1];

            return ({
                x: {
                    type: 'time' as const,
                    nice: true
                },
                y: {
                    type: 'linear' as const,
                    domain: [firstTickValue, lastTickValue],
                    nice: false,
                    zero: false,
                    round: false,
                    clamp: true,
                }
            })
        },
        [accessors, sortedData, numberOfTicksY]
    );

    const xScale = useMemo(
        () => scaleTime({
                nice: true,
                range: [0, innerWidth],
                domain: [accessors.x(sortedData[0]), accessors.x(sortedData[sortedData.length - 1])]
            }),
        [accessors, sortedData, innerWidth]
    );

    const yScale = useMemo(
        () => scaleLinear({...scalesConfig.y, range: [innerHeight, 0]}),
        [scalesConfig.y, innerHeight]
    );

    // state
    const [activeDatum, setActiveDatum] = useState<EventHandlerParams<any> & { line: LineChartProps['series'][number] } | null>(null);

    // callbacks
    const renderTooltip = useCallback(
        (renderProps: RenderTooltipParams<any>) =>
            <TooltipContent
                {...renderProps}
                accessors={accessors}
                series={series}
                className='line-chart__tooltip-content'/>,
        [accessors, series]
    );

    // Because xychart fires all consumer's handlers twice, we need to debounce our handler
    // Remove debounce when https://github.com/airbnb/visx/issues/1077 will be resolved
    const handlePointerUp = useDebouncedCallback(
        (event: EventHandlerParams<any>) => {
            const line = series.find(line => line.dataKey === event.key);

            // By types from visx/xychart index can be undefined
            const activeDatumIndex = activeDatum?.index;

            if (!event.event || !line || !isValidNumber(activeDatumIndex)) {
                return;
            }

            onDatumClick({
                originEvent: event.event as MouseEvent<unknown>,
                link: line?.linkURLs?.[activeDatumIndex],
            });
        },
    );

    const handlePointerMove = useDebouncedCallback(
        (event: EventHandlerParams<any>) => {
            const line = series.find(line => line.dataKey === event.key);

            if (!line) {
                return;
            }

            setActiveDatum({
                ...event,
                line
            });
        },
        0,
        { leading: true }
    );

    const handlePointerOut = useDebouncedCallback(
        () => setActiveDatum(null)
    );

    const activeDatumLink = activeDatum?.line?.linkURLs?.[activeDatum?.index];

    const rootClasses = classnames(
        'line-chart__content',
        { 'line-chart__content--with-cursor': !!activeDatumLink }
    );

    return (
        <div className={rootClasses}>
            <XYChart
                theme={lightTheme}
                xScale={scalesConfig.x}
                yScale={scalesConfig.y}
                height={height}
                width={width}
                captureEvents={true}
                margin={MARGIN}
                onPointerMove={handlePointerMove}
                onPointerUp={handlePointerUp}
                onPointerOut={handlePointerOut}
            >

                <Group top={MARGIN.top} left={MARGIN.left}>

                    <GridRows
                        scale={yScale}
                        numTicks={numberOfTicksY}
                        width={innerWidth}
                        className='line-chart__grid-line'
                    />

                    <GridColumns
                        scale={xScale}
                        numTicks={numberOfTicks}
                        height={innerHeight}
                        className='line-chart__grid-line'
                    />
                </Group>

                <Axis
                    orientation="bottom"
                    tickValues={xScale.ticks(numberOfTicks)}
                    tickFormat={formatDate}
                    numTicks={numberOfTicks}
                    axisClassName='line-chart__axis'
                    axisLineClassName='line-chart__axis-line'
                    tickClassName="line-chart__axis-tick"
                />
                <Axis
                    orientation="left"
                    numTicks={numberOfTicksY}
                    tickFormat={format('~s')}
                    axisClassName='line-chart__axis'
                    axisLineClassName='line-chart__axis-line'
                    tickClassName="line-chart__axis-tick"
                />

                {
                    series.map(line =>
                        <Group key={line.dataKey as string}>

                            <LineSeries
                                dataKey={line.dataKey as string}
                                data={sortedData}
                                strokeWidth={3}
                                xAccessor={accessors.x}
                                yAccessor={accessors.y[line.dataKey as string]}
                                stroke={line.stroke ?? lightTheme.colors[0]}
                                curve={curveLinear}
                            />

                            <GlyphSeries
                                dataKey={line.dataKey as string}
                                data={sortedData}
                                /* eslint-disable-next-line react/jsx-no-bind */
                                colorAccessor={() => line.stroke}
                                xAccessor={accessors.x}
                                yAccessor={accessors.y[line.dataKey as string]}
                                renderGlyph={GlyphDotComponent}
                            />
                        </Group>
                    )
                }

                <Group top={MARGIN.top} left={MARGIN.left}>
                    {
                        activeDatum &&
                        <GlyphDot
                            className='line-chart__glyph line-chart__glyph--active'
                            r={8}
                            fill={activeDatum.line.stroke}
                            cx={xScale(accessors.x(activeDatum.datum))}
                            cy={yScale(accessors.y[activeDatum.key](activeDatum.datum))}
                        />
                    }
                </Group>

                <Tooltip
                    className='line-chart__tooltip'
                    showHorizontalCrosshair={false}
                    showVerticalCrosshair={true}
                    snapTooltipToDatumX={false}
                    snapTooltipToDatumY={false}
                    showDatumGlyph={false}
                    showSeriesGlyphs={false}
                    renderTooltip={renderTooltip}
                />
            </XYChart>
        </div>
    );
}

export function LineChart(props: LineChartProps): ReactElement {

    const { width, height, ...otherProps } = props;
    const hasLegend = props.series.every(line => !!line.name);

    if (!hasLegend) {
        return (<LineChartContentComponent {...props}/>)
    }

    return (
        /* eslint-disable-next-line react/forbid-dom-props */
        <div style={{ width, height }} className='line-chart'>
            {/*
                In case if we have a legend to render we have to have responsive container for chart
                just to calculate right sizes for chart content = rootContainerSizes - legendSizes
            */}
            <ParentSize className='line-chart__content-parent-size'>
                {
                    ({ width, height}) => (<LineChartContentComponent {...otherProps} width={width} height={height}/>)
                }
            </ParentSize>

            <ul className='line-chart__legend'>

                { props.series.map(line => (
                        <li key={line.dataKey.toString()} className='line-chart__legend-item'>

                            {/* eslint-disable-next-line react/forbid-dom-props */}
                            <div style={{ backgroundColor: line.stroke }} className='line-chart__legend-mark' />
                            {line.name}
                        </li>
                    ))}
            </ul>
        </div>
    )
}
