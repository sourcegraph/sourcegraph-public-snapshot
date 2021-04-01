import React, { ReactElement, useCallback, useMemo } from 'react';
import { LineChartContent } from 'sourcegraph';
import { curveLinear } from '@visx/curve';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';
import {
    Axis,
    EventEmitterProvider,
    GlyphSeries,
    Grid,
    lightTheme,
    LineSeries,
    Tooltip,
    XYChart,
} from '@visx/xychart';
import useEventEmitters from '@visx/xychart/lib/hooks/useEventEmitters';

import { XYCHART_EVENT_SOURCE } from '@visx/xychart/lib/constants';
import { format } from 'd3-format';
import { timeFormat } from 'd3-time-format'
import { scaleLinear, scaleTime } from '@visx/scale';
import { GridColumns } from '@visx/grid'
import { Group } from '@visx/group'
import { TextProps } from '@visx/text/lib/Text'

import { generateAccessors } from './helpers/generate-accessors'
import { getRangeWithPadding } from './helpers/get-range-with-padding'
import { getMinAndMax } from './helpers/get-min-max'
import { GlyphComponent } from './components/Glyph'
import { TooltipContent } from './components/TooltipContent'
import { MaybeLink } from './components/MaybeLink'

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
    onDataPointClick?: () => void;
}

export function LineChart(props: LineChartProps): ReactElement {
    return (
        <EventEmitterProvider>
            <LineChartContentComponent {...props}/>
        </EventEmitterProvider>
    )
}

function LineChartContentComponent(props: LineChartProps): ReactElement {
    const { width, height, data, series, xAxis, onDataPointClick = () => {} } = props;

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

    // callbacks
    const renderTooltip = useCallback(
        (renderProps: RenderTooltipParams<any>) =>
            <TooltipContent {...renderProps} accessors={accessors} series={series}/>,
        [accessors, series]
    );

    const eventEmitters = useEventEmitters({ source: XYCHART_EVENT_SOURCE });

    return (

        <XYChart
            theme={lightTheme}
            xScale={scalesConfig.x}
            yScale={scalesConfig.y}
            height={height}
            width={width}
            captureEvents={false}
            margin={MARGIN}
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
                    <LineSeries
                        key={line.dataKey as string}
                        dataKey={line.dataKey as string}
                        data={sortedData}
                        strokeWidth={3}
                        xAccessor={accessors.x}
                        yAccessor={accessors.y[line.dataKey as string]}
                        stroke={line.stroke ?? lightTheme.colors[0]}
                        curve={curveLinear}
                    />
                )
            }

            <Tooltip
                debounce={200}
                showHorizontalCrosshair={false}
                showVerticalCrosshair={true}
                snapTooltipToDatumX={false}
                snapTooltipToDatumY={true}
                showDatumGlyph={true}
                showSeriesGlyphs={true}
                renderTooltip={renderTooltip}
            />

            <rect
                x={MARGIN.left}
                y={MARGIN.top}
                width={width - MARGIN.left - MARGIN.right}
                height={height - MARGIN.top - MARGIN.bottom}
                fill="transparent"
                {...eventEmitters}
            />

            {
                series.map(line =>
                    <GlyphSeries
                        key={line.dataKey as string}
                        dataKey={line.dataKey as string}
                        data={sortedData}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        colorAccessor={() => line.stroke}
                        xAccessor={accessors.x}
                        yAccessor={accessors.y[line.dataKey as string]}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        renderGlyph={props => (
                            <MaybeLink
                                // visx types are wrong here. props don't have index value
                                // key is index here. Index doesn't exist in runtime
                                href={line.linkURLs?.[+props.key]}
                                onClick={onDataPointClick}
                                {...eventEmitters}
                            >

                                <GlyphComponent {...props}/>
                            </MaybeLink>
                        )}
                    />
                )
            }
        </XYChart>
    );
}
