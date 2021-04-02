import React, { ReactElement, useCallback, useMemo, useState, MouseEvent } from 'react';
import classnames from 'classnames';
import { LineChartContent } from 'sourcegraph';
import { curveLinear } from '@visx/curve';
import { useDebouncedCallback } from 'use-debounce';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';
import {
    Axis,
    GlyphSeries,
    Grid,
    lightTheme,
    LineSeries,
    Tooltip,
    XYChart,
} from '@visx/xychart';

import { format } from 'd3-format';
import { timeFormat } from 'd3-time-format'
import { scaleLinear, scaleTime } from '@visx/scale';
import { GridColumns } from '@visx/grid'
import { Group } from '@visx/group'
import { TextProps } from '@visx/text/lib/Text'
import { EventHandlerParams } from '@visx/xychart/lib/types';

import { generateAccessors } from './helpers/generate-accessors'
import { getRangeWithPadding } from './helpers/get-range-with-padding'
import { getMinAndMax } from './helpers/get-min-max'
import { GlyphComponent } from './components/Glyph'
import { TooltipContent } from './components/TooltipContent'
import { onDatumClick } from '../types';

// Chart configuration
const WIDTH_PER_TICK = 70;
const HEIGHT_PER_TICK = 40;
const MARGIN = { top: 10, left: 30, bottom: 20, right: 20 };
const TICK_LABEL_PROPS = (): Partial<TextProps> => ({
    fill: 'black',
    fontSize: 12,
    fontWeight: 400
})

// Formatters
const dateFormatter = timeFormat('%d %b');
const formatDate = (date: Date): string => dateFormatter(date);

export interface LineChartProps extends Omit<LineChartContent<any, string>, 'chart'> {
    width: number;
    height: number;
    onDatumClick: onDatumClick;
}

export function LineChart(props: LineChartProps): ReactElement {
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

    // state
    const [activeLinkDatum, setActiveLinkDatum] = useState<any | null>(null);

    // callbacks
    const renderTooltip = useCallback(
        (renderProps: RenderTooltipParams<any>) =>
            <TooltipContent {...renderProps} accessors={accessors} series={series}/>,
        [accessors, series]
    );

    // Because xychart fires all consumer's handlers twice, we need to debounce our handler
    // Remove debounce when https://github.com/airbnb/visx/issues/1077 will be resolved
    const handlePointerUp = useDebouncedCallback(
        (event: EventHandlerParams<any>) => {

            if (!event.event) {
                return;
            }

            onDatumClick({
                originEvent: event.event as MouseEvent<unknown>,
                link: activeLinkDatum
            });
        },
    );

    const handlePointerMove = useDebouncedCallback(
        (event: EventHandlerParams<any>) => {
            const line = series.find(line => line.dataKey === event.key);

            if (!line) {
                return;
            }

            const link = line?.linkURLs?.[event.index];

            setActiveLinkDatum(link);
        },
        0,
        { leading: true }
    );

    return (
        <div className={classnames('line-chart', { 'line-chart--with-cursor': !!activeLinkDatum })}>

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
            >

                <Group top={MARGIN.top} left={MARGIN.left}>

                    <GridColumns
                        scale={xScale}
                        numTicks={numberOfTicks}
                        width={innerWidth} height={innerHeight} stroke="#e0e0e0"
                        lineStyle={{ stroke: 'gray', strokeWidth: 1, strokeOpacity: 0.3 }} />
                </Group>

                <Grid
                    rows={true}
                    columns={false}
                    numTicks={numberOfTicksY}
                    lineStyle={{ stroke: 'gray', strokeWidth: 1, strokeOpacity: 0.3 }}
                />

                <Axis
                    orientation="bottom"
                    strokeWidth={2}
                    stroke="black"
                    tickStroke="black"
                    tickClassName="ticks"
                    tickValues={xScale.ticks(numberOfTicks)}
                    tickFormat={formatDate}
                    numTicks={numberOfTicks}
                    tickLabelProps={TICK_LABEL_PROPS}
                />
                <Axis
                    orientation="left"
                    numTicks={numberOfTicksY}
                    strokeWidth={2}
                    stroke="black"
                    tickFormat={format('~s')}
                    tickStroke="black"
                    tickClassName="ticks"
                    tickLabelProps={TICK_LABEL_PROPS}
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
                                renderGlyph={GlyphComponent}
                            />
                        </Group>
                    )
                }

                <Tooltip
                    showHorizontalCrosshair={false}
                    showVerticalCrosshair={true}
                    snapTooltipToDatumX={false}
                    snapTooltipToDatumY={false}
                    showDatumGlyph={true}
                    showSeriesGlyphs={true}
                    renderTooltip={renderTooltip}
                />
            </XYChart>
        </div>
    );
}
