import React, { MouseEventHandler, ReactElement, useCallback } from 'react';
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

    const content = pies[0];
    const { data, dataKey, nameKey, linkURLKey = '', fillKey = '' } = content;

    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const radius = Math.min(innerWidth, innerHeight) / 3;
    const centerY = innerHeight / 2;
    const centerX = innerWidth / 2;

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
        <svg width={width} height={height}>
            <Group top={centerY + margin.top} left={centerX + margin.left}>
                <Pie
                    data={data}
                    pieValue={usage}
                    outerRadius={radius}
                    cornerRadius={3}
                    padRadius={30}
                >
                    {pie => (
                        <Group>
                            {
                                pie.arcs.map(arc => {
                                        const link = getLink(arc);
                                        const classes = classnames('pie-chart__arc', { 'pie-chart__arc--with-link': link })

                                        return (
                                            <PieArc
                                                key={getKey(arc)}
                                                className={classes}
                                                arc={arc}
                                                path={pie.path}
                                                getColor={getFillColor}
                                                getKey={getKey}
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
}

function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { path, arc, getColor, getKey, className, onClick } = props;
    const pathValue = path(arc) ?? '';
    const middleAngle = Math.PI / 2 - (arc.startAngle + ((arc.endAngle - arc.startAngle) / 2));

    const radius = path.outerRadius()(arc);
    const normalX = Math.cos(middleAngle);
    const normalY = Math.sin(-middleAngle);

    const labelX = normalX * (30);
    const labelY = normalY * (30);

    const surfaceX = normalX * (radius + 2);
    const surfaceY = normalY * (radius + 2);

    return (
        <Group
            className={className}
            onClick={onClick}>

            <path
                d={pathValue}
                fill={getColor(arc)}
                stroke='white'
                strokeWidth={1}
            />
            <Annotation
                x={surfaceX}
                y={surfaceY}
                dx={labelX}
                dy={labelY}
            >
                <Connector type='line' />
                <Label
                    showBackground={false}
                    showAnchorLine={false}
                    title={getKey(arc)}
                    subtitle={`${arc.value}%`} />
            </Annotation>

            <circle r={4} fill="black" cx={surfaceX + labelX} cy={surfaceY + labelY}/>
        </Group>
    );
}
