import React, { MouseEventHandler, PointerEventHandler, ReactElement, useCallback, useMemo, useState } from 'react';
import classnames from 'classnames';
import { PieChartContent } from 'sourcegraph';
import Pie, { PieArcDatum } from '@visx/shape/lib/shapes/Pie';
import { Group } from '@visx/group';
import { Annotation, Connector } from '@visx/annotation';
import { Arc as ArcType } from 'd3-shape';

// Replace import below on standard @visx/annotation package
// when ticket about bad label positioning will be resolve
// https://github.com/airbnb/visx/issues/1126
import { Label } from '../../annotation/Label';
import { onDatumClick } from '../types';

const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 20, left: 20 };

export interface PieChartProps extends PieChartContent<any> {
    width: number;
    height: number;
    margin?: typeof DEFAULT_MARGIN;
    onDatumClick: onDatumClick
}

export function PieChart(props: PieChartProps): ReactElement | null {
    const {
        width,
        height,
        margin = DEFAULT_MARGIN,
        pies,
        onDatumClick,
    } = props;

    const [activeArc, setActiveArc] = useState<PieArcDatum<any> | null>(null)

    const content = pies[0];
    const { data, dataKey, nameKey, linkURLKey = '', fillKey = '' } = content;

    const sortedData = useMemo(
        () => {
            const sortedData = [...data.sort(
                (first, second) => +first[dataKey] - +second[dataKey])
            ];

            const result = [];

            while (sortedData.length) {
                const firstElement = sortedData.shift();

                if (firstElement) {
                    result.push(firstElement)
                }

                const lastElement = sortedData.pop();

                if (lastElement) {
                    result.push(lastElement);
                }
            }

            return result;
        },
        [data, dataKey]
    )

    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const radius = Math.min(innerWidth, innerHeight) / 4;
    const centerY = innerHeight / 2;
    const centerX = innerWidth / 2;

    const sum = useMemo(
        () => sortedData.reduce((sum, datum) => +sum + datum[dataKey], 0),
        [sortedData, dataKey]
    );

    // accessor
    const usage = useCallback(
        (data: any): number => data[dataKey],
        [dataKey]
    );

    const getKey = useCallback(
        (arc: PieArcDatum<any>): string => arc.data[nameKey],
        [nameKey]
    );

    const getFillColor = useCallback(
        (arc: PieArcDatum<any>): string => arc.data[fillKey] ?? 'grayscale',
        [fillKey]
    );

    const getLink = useCallback(
        (arc: PieArcDatum<any>): string => arc.data[linkURLKey],
        [linkURLKey]
    );

    if (width < 10) {
        return null;
    }

    return (
        <svg className='pie-chart' width={width} height={height}>

            <Group top={centerY + margin.top} left={centerX + margin.left}>

                <Pie
                    data={sortedData}
                    pieValue={usage}
                    outerRadius={radius}
                    cornerRadius={3}
                    pieSort={null}
                    pieSortValues={null}
                    padRadius={40}
                >
                    {pie => (
                        <Group>
                            {
                                pie.arcs.map(arc => {
                                        const link = getLink(arc);
                                        const classes = classnames(
                                            'pie-chart__arc',
                                            { 'pie-chart__arc--with-link': link, 'pie-chart__arc--active': activeArc && getKey(arc) === getKey(activeArc) })

                                        return (
                                            <PieArc
                                                key={getKey(arc)}
                                                className={classes}
                                                arc={arc}
                                                path={pie.path}
                                                sum={sum}
                                                getColor={getFillColor}
                                                getKey={getKey}
                                                /* eslint-disable-next-line react/jsx-no-bind */
                                                onPointerMove={() => setActiveArc(arc)}
                                                /* eslint-disable-next-line react/jsx-no-bind */
                                                onClick={event =>
                                                    onDatumClick({ originEvent: event, link })
                                                }
                                            />
                                        );
                                    }
                                )
                            }

                            {
                                pie.arcs.map(arc => {
                                        const link = getLink(arc);
                                        const classes = classnames(
                                            'pie-chart__arc',
                                            {
                                                'pie-chart__arc--with-link': link,
                                                'pie-chart__arc--with-annotation-only': !(activeArc && getKey(arc) === getKey(activeArc))
                                            }
                                        );

                                        return (
                                            <PieArc
                                                key={getKey(arc)}
                                                className={classes}
                                                arc={arc}
                                                path={pie.path}
                                                sum={sum}
                                                getColor={getFillColor}
                                                getKey={getKey}
                                                /* eslint-disable-next-line react/jsx-no-bind */
                                                onPointerMove={() => setActiveArc(arc)}
                                                /* eslint-disable-next-line react/jsx-no-bind */
                                                onClick={event =>
                                                    onDatumClick({ originEvent: event, link })
                                                }
                                            />
                                        );
                                    }
                                )
                            }
                        </Group>
                    )}
                </Pie>
            </Group>
        </svg>
    );
}

// Components helpers

interface PieArcProps<Datum> {
    className?: string;
    getKey: (d: PieArcDatum<Datum>) => string;
    getColor: (d: PieArcDatum<Datum>) => string;
    path: ArcType<any, PieArcDatum<Datum>>;
    arc: PieArcDatum<Datum>;
    onClick: MouseEventHandler
    sum: number;
    onPointerMove?: PointerEventHandler;
    onPointerOut?: PointerEventHandler;
}

function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { sum, path, arc, getColor, getKey, className, onClick, onPointerMove, onPointerOut } = props;
    const pathValue = path(arc) ?? '';
    const middleAngle = Math.PI / 2 - (arc.startAngle + ((arc.endAngle - arc.startAngle) / 2));

    const radius = path.outerRadius()(arc);
    const normalX = Math.cos(middleAngle);
    const normalY = Math.sin(-middleAngle);

    const labelX = normalX * (15);
    const labelY = normalY * (15);

    const surfaceX = normalX * (radius + 2);
    const surfaceY = normalY * (radius + 2);

    return (
        <Group
            className={className}
            onClick={onClick}
            onPointerMove={onPointerMove}
            onPointerOut={onPointerOut}
        >

            <path
                className='pie-chart__arc-path'
                d={pathValue}
                fill={getColor(arc)}
            />

            <Annotation
                x={surfaceX}
                y={surfaceY}
                dx={labelX}
                dy={labelY}
            >
                <Connector
                    className='pie-chart__label-line'
                    type='line' />

                <Label
                    className='pie-chart__label'
                    backgroundPadding={4}
                    showBackground={true}
                    showAnchorLine={false}
                    title={getKey(arc)}
                    subtitle={`${(100 * arc.value / sum).toFixed(2)}%`}/>
            </Annotation>

            <circle
                className='pie-chart__label-circle'
                r={2}
                fill="black"
                cx={surfaceX + labelX}
                cy={surfaceY + labelY}/>

        </Group>
    );
}
