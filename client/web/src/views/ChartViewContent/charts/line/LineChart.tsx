import React, {ReactElement, useCallback, useMemo} from 'react';
import { ChartAxis, LineChartContent } from 'sourcegraph';
import { curveLinear } from '@visx/curve';
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip';

import {
    Axis,
    Grid,
    LineSeries,
    Tooltip,
    XYChart,
    lightTheme, GlyphProps, GlyphSeries,
} from '@visx/xychart';
import {GlyphDot} from '@visx/glyph';

export interface XYChartProps extends LineChartContent<any, string> {
    width: number;
    height: number;
}

// Chart configuration
// const glyphComponent: 'star' | 'cross' | 'circle' = 'circle';
const glyphOutline = 'white'; // any color

const TICK_LABEL_PROPS = (): object => ({ fill: 'black' })
const NUMBER_OF_TICKS = 4;
const MARGIN = { top: 10, left: 50, bottom: 30, right: 30 };

// Helpers
export function pad([min, max]:number[], coefficient: number): [number, number] {
    const increment = (max - min) * coefficient / 2;

    return [min - increment, max + increment];
}

interface Accessors<Datum, Key extends string | number> {
    x: (d: Datum) => Date | number;
    y: Record<Key, (data: Datum) => any>;
}

function generateAccessors<Datum extends object>(
    xAxis: ChartAxis<keyof Datum, Datum>,
    series: { dataKey: keyof Datum}[]
): Accessors<Datum, string> {
    const { dataKey: xDataKey, scale = 'time' } = xAxis;

    return {
        x: data => scale === 'time'
            // as unknown as string quick hack for cast Datum[keyof Datum] to string
            // fix that when we will have a value type for LineChartContent<D> generic
            ? new Date(data[xDataKey] as unknown as string)
            // In case if we got linear scale we have to operate with numbers
            : +data[xDataKey],
        y: series.reduce<Record<string, (data: Datum) => any>>((accessors, currentLine) => {
            const { dataKey } = currentLine;
            // as unknown as string quick hack for cast Datum[keyof Datum] to string
            // fix that when we will have a value type for LineChartContent<D> generic
            const key = dataKey as unknown as string;

            accessors[key] = data => +data[dataKey];

            return accessors;
        }, {})
    };
}

function getMinAndMax<Datum, Key extends string | number>(data: Datum[], accessors: Accessors<Datum, Key> ): [number, number] {
    const keys = Object.keys(accessors.y) as Key[];

    const resultArray = data.reduce<number[]>((memo, item) => {
        for (const key of keys) {
            const accessor = accessors.y[key];

            memo.push(+accessor(item))
        }

        return memo;
    }, []);

    return [Math.min(...resultArray), Math.max(...resultArray)]
}

function GlyphComponent(props: GlyphProps<any>): ReactElement {
    const { x: xCoord, y: yCoord, color, onPointerMove, onPointerOut, onPointerUp } = props;
    const handlers = { onPointerMove, onPointerOut, onPointerUp };

    return (
        <GlyphDot
            cx={xCoord}
            cy={yCoord}
            stroke={glyphOutline}
            strokeWidth={2}
            fill={color}
            r={4}
            {...handlers}
        />
    );
}

export function LineChart(props: XYChartProps) {
    const { width, height, data, series, xAxis } = props;

    // derived
    const accessors = useMemo(() => generateAccessors(xAxis, series), [xAxis, series])
    const scalesConfig = useMemo(() => ({
        x: {
            type: 'time' as const,
            paddingInner: 0.3,
            nice: false
        },
        y: {
            type: 'linear' as const,
            domain: pad(getMinAndMax(data, accessors), 0.3),
            nice: true,
            zero: true,
        }
    }), [accessors, data])

    // callbacks
    const renderTooltip = useCallback(({ tooltipData, colorScale }: RenderTooltipParams<any>) => (
        <>
            {/** date */}
            {(tooltipData?.nearestDatum?.datum &&
                new Date(accessors.x(tooltipData?.nearestDatum?.datum)).toDateString()) ||
            'No date'}
            <br/>
            <br/>
            {/** values */}
            {(Object.keys(tooltipData?.datumByKey ?? {}).filter(lineKey => lineKey) as any[]).map(lineKey => {
                const value =
                    tooltipData?.nearestDatum?.datum &&
                    accessors.y[lineKey](
                        tooltipData?.nearestDatum?.datum,
                    );

                const line = series.find(line => line.dataKey === lineKey)

                /* eslint-disable react/forbid-dom-props */
                return (
                    <div
                        className='line-chart__tooltip'
                        key={lineKey}>

                        <em
                            className='line-chart__tooltip-text'
                            style={{
                                color: colorScale?.(lineKey),
                                textDecoration:
                                    tooltipData?.nearestDatum?.key === lineKey ? 'underline' : undefined,
                            }}
                        >
                            {line?.name ?? 'unknown series'}
                        </em>{' '}
                        {
                          value === null || Number.isNaN(value)
                            ? '–'
                            : value
                        }
                    </div>
                );
            })}
        </>
    ), [accessors, series]);

    return (
        <XYChart
            theme={lightTheme}
            xScale={scalesConfig.x}
            yScale={scalesConfig.y}
            height={height}
            width={width}
            captureEvents={true}
            margin={MARGIN}
        >
            <Grid
                rows={true}
                columns={true}
                lineStyle={{ stroke: 'gray', strokeWidth: 1, strokeOpacity: 0.2 }}
            />

            <Axis
                orientation="bottom"
                strokeWidth={2}
                stroke="black"
                tickStroke="black"
                tickClassName="ticks"
                tickLabelProps={TICK_LABEL_PROPS}
            />
            <Axis
                orientation="left"
                numTicks={NUMBER_OF_TICKS}
                strokeWidth={2}
                stroke="black"
                tickStroke="black"
                tickClassName="ticks"
                tickLabelProps={TICK_LABEL_PROPS}
            />

            {
                series.map(line =>
                    <g key={line.dataKey as string}>
                        <LineSeries
                            dataKey={line.dataKey as string}
                            data={data}
                            xAccessor={accessors.x}
                            yAccessor={accessors.y[line.dataKey as string]}
                            curve={curveLinear}
                        />

                        <GlyphSeries
                            dataKey={line.dataKey as string}
                            data={data}
                            xAccessor={accessors.x}
                            yAccessor={accessors.y[line.dataKey as string]}
                            renderGlyph={GlyphComponent}
                        />
                    </g>
                )
            }

            <Tooltip
                showHorizontalCrosshair={false}
                showVerticalCrosshair={true}
                snapTooltipToDatumX={false}
                snapTooltipToDatumY={true}
                showDatumGlyph={true}
                showSeriesGlyphs={true}
                renderTooltip={renderTooltip}
            />
        </XYChart>
    );
}
