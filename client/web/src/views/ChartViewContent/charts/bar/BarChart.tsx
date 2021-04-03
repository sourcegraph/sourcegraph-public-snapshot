import React, { ReactElement, useCallback, useMemo } from 'react';
import classnames from 'classnames';
import { scaleBand, scaleLinear } from '@visx/scale';
import { AxisBottom, AxisLeft } from '@visx/axis'
import { localPoint } from '@visx/event';
import { Group} from '@visx/group';
import { Bar} from '@visx/shape';
import { GridRows } from '@visx/grid';
import { useTooltip, TooltipWithBounds } from '@visx/tooltip';
import { BarChartContent } from 'sourcegraph';

import { onDatumClick } from '../types';

const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 25, left: 40 };

// Tooltip timeout used below as semaphore to prevent tooltip flashing
// in case if user is moving mouse fast
let tooltipTimeout: number;


// helpers
// eslint-disable-next-line id-length
const range = (rangeLength: number): number[] => [...new Array(rangeLength)].map((_, index) => index);

type Accessor<Datum> = (d: Datum) => number | string;

// Compose together the scale and accessor functions to get point functions
function compose<Datum>(scale: any, accessor: Accessor<Datum>) { return  (data: Datum) => scale(accessor(data));}

interface TooltipData {
    xLabel: string;
    value: number;
}

interface BarChartProps extends Omit<BarChartContent<any, string>, 'chart'> {
    width: number;
    height: number;
    onDatumClick: onDatumClick
}

export function BarChart(props: BarChartProps): ReactElement {

    const { width, height, data, series, onDatumClick } = props;

    // Respect only first element of data series
    // Refactor this in case if we need support stacked bar chart
    const { dataKey, fill, linkURLs } = series[0];

    const xMax = width - DEFAULT_MARGIN.left - DEFAULT_MARGIN.right;
    const yMax = height - DEFAULT_MARGIN.top - DEFAULT_MARGIN.bottom;

    const {
        tooltipOpen,
        tooltipLeft,
        tooltipTop,
        tooltipData,
        hideTooltip,
        showTooltip,
    } = useTooltip<TooltipData>();

    // Accessors
    const yAccessor = useCallback(
        (data: any) => data[dataKey],
        [dataKey]
    );

    const formatXLabel = useCallback(
        index => data[index].name,
        [data]
    );

    // And then scale the graph by our data
    const xScale = useMemo(() =>
            scaleBand({
                range: [0, xMax],
                round: true,
                domain: range(data.length),
                padding: 0.2,
            }),
        [xMax, data]
    );

    const yScale = useMemo(() =>
            scaleLinear({
                range: [yMax, 0],
                round: true,
                nice: true,
                domain: [0, Math.max(...data.map(yAccessor))],
            }),
        [yMax, data, yAccessor]
    );

    const yPoint = useMemo(() => compose(yScale, yAccessor), [yScale, yAccessor]);

    // handlers
    const handleMouseLeave = useCallback(
        () => {
            tooltipTimeout = window.setTimeout(() => {
                hideTooltip();
            }, 300);
        },
        [hideTooltip]
    );

    return (
        <div className='bar-chart'>
            <svg width={width} height={height}>
                <Group left={DEFAULT_MARGIN.left} top={DEFAULT_MARGIN.top}>

                    <GridRows
                        scale={yScale}
                        width={xMax}
                        height={yMax}
                        className='bar-chart__grid'
                    />

                    {
                        data.map((datum, index) => {
                            const barHeight = yMax - (yPoint(datum) ?? 0);
                            const link = linkURLs?.[index];
                            const classes =  classnames(
                                'bar-chart__bar',
                                { 'bar-chart__bar--with-link': link }
                            );

                            return (
                                <Group key={`bar-${index}`}>

                                    <Bar
                                        className={classes}
                                        x={xScale(index)}
                                        y={yMax - barHeight}
                                        height={barHeight}
                                        width={xScale.bandwidth()}
                                        fill={fill}
                                        /* eslint-disable-next-line react/jsx-no-bind */
                                        onClick={event => {
                                            const link = linkURLs?.[index];

                                            onDatumClick({ originEvent: event, link })
                                        }}
                                        onMouseLeave={handleMouseLeave}

                                        // In this case we have to use arrow function because we need
                                        // get access to index and current datum within onMouseMove handler
                                        /* eslint-disable-next-line react/jsx-no-bind */
                                        onMouseMove={event => {
                                            if (tooltipTimeout) { clearTimeout(tooltipTimeout); }

                                            const rectangle = localPoint(event);

                                            showTooltip({
                                                tooltipData: { xLabel: formatXLabel(index), value: yAccessor(datum) },
                                                tooltipTop: rectangle?.y,
                                                tooltipLeft:  rectangle?.x,
                                            });
                                        }}
                                    />
                                </Group>
                            );
                        })
                    }

                    <AxisBottom
                        top={yMax}
                        scale={xScale}
                        tickFormat={formatXLabel}
                        axisClassName='bar-chart__axis'
                        axisLineClassName='bar-chart__axis-line'
                        tickClassName="bar-chart__axis-tick"
                    />

                    <AxisLeft
                        scale={yScale}
                        axisClassName='bar-chart__axis'
                        axisLineClassName='bar-chart__axis-line'
                        tickClassName="bar-chart__axis-tick"
                    />
                </Group>
            </svg>

            {tooltipOpen && tooltipData && (
                <TooltipWithBounds
                    className='bar-chart__tooltip'
                    top={tooltipTop}
                    left={tooltipLeft}>

                    <div className='bar-chart__tooltip-content'>

                        <strong className='bar-chart__tooltip-name'>
                            {tooltipData.xLabel}
                        </strong>
                    </div>

                    <div className='bar-chart__tooltip-value'>
                        {tooltipData.value}
                    </div>
                </TooltipWithBounds>
            )}
        </div>
    );
}
