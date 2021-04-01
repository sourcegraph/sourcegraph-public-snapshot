import React, { ReactElement, useCallback, useMemo } from 'react';
import {scaleBand, scaleLinear} from '@visx/scale';
import {AxisBottom, AxisLeft} from '@visx/axis'
import { localPoint } from '@visx/event';
import {Group} from '@visx/group';
import {Bar} from '@visx/shape';
import {Grid} from '@visx/grid';
import {useTooltip, defaultStyles, TooltipWithBounds} from '@visx/tooltip';
import {BarChartContent} from 'sourcegraph';
import {TextProps} from '@visx/text/lib/Text';

const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 20, left: 40 };

// Tooltip timeout used below as semaphore to prevent tooltip flashing
// in case if user is moving mouse fast
let tooltipTimeout: number;

const tooltipStyles = {
    ...defaultStyles,
    minWidth: 60,
    backgroundColor: 'rgba(0,0,0,0.9)',
    color: 'white',
};

const tickLabelStyles = (): Partial<TextProps> => ({
    fill: 'black',
    fontSize: 11,
    textAnchor: 'middle',
})

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

interface BarChartProps extends BarChartContent<any, string> {
    width: number;
    height: number;
}

export function BarChart(props: BarChartProps): ReactElement {

    const { width, height, data, series } = props;

    // Respect only first element of data series
    // Refactor this in case if we need support stacked bar chart
    const { dataKey, fill, linkURLs, name } = series[0];

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

                    <Grid
                        xScale={xScale}
                        yScale={yScale}
                        width={xMax}
                        height={yMax}
                        stroke="black"
                        strokeOpacity={0.1}
                        xOffset={xScale.bandwidth() / 2}
                    />

                    {
                        data.map((datum, index) => {
                            const barHeight = yMax - (yPoint(datum) ?? 0);

                            return (
                                <Group key={`bar-${index}`}>

                                    <Bar
                                        x={xScale(index)}
                                        y={yMax - barHeight}
                                        height={barHeight}
                                        width={xScale.bandwidth()}
                                        fill={fill}
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
                        stroke='black'
                        tickStroke='black'
                        tickLabelProps={tickLabelStyles}
                    />

                    <AxisLeft scale={yScale} />
                </Group>
            </svg>

            {tooltipOpen && tooltipData && (
                <TooltipWithBounds
                    top={tooltipTop}
                    left={tooltipLeft}
                    style={tooltipStyles}>

                    <div>
                        <strong>{tooltipData.xLabel}</strong>
                    </div>
                    <div>{tooltipData.value}</div>
                </TooltipWithBounds>
            )}
        </div>
    );
}
