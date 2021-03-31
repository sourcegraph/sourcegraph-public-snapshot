import React, {ReactElement, useCallback} from 'react';
import { PieChartContent } from 'sourcegraph';
import Pie, { ProvidedProps, PieArcDatum } from '@visx/shape/lib/shapes/Pie';
import { Group } from '@visx/group';
import { Annotation, Connector } from '@visx/annotation';
import { Arc as ArcType } from 'd3-shape';

// Replace import below on standard @visx/annotation package
// when ticket about bad label positioning will be resolve
// https://github.com/airbnb/visx/issues/1126
import { Label } from '../../annotation/Label';

const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 20, left: 20 };

export interface PieChartProps extends PieChartContent<any> {
    width: number;
    height: number;
    margin?: typeof DEFAULT_MARGIN;
}

export function PieChart(props: PieChartProps): ReactElement | null {
    const {
        width,
        height,
        margin = DEFAULT_MARGIN,
        pies
    } = props;

    const content = pies[0];
    const { data, dataKey, nameKey, fillKey = '' } = content;

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
                        <PieArcs
                            {...pie}
                            getKey={getKey}
                            getColor={getFillColor}
                        />
                    )}
                </Pie>
            </Group>
        </svg>
    );
}

// Components helpers
type AnimatedPieProps<Datum> = ProvidedProps<Datum> & {
    getKey: (d: PieArcDatum<Datum>) => string;
    getColor: (d: PieArcDatum<Datum>) => string;
    onClickDatum?: (d: PieArcDatum<Datum>) => void;
};

function PieArcs<Datum>(props: AnimatedPieProps<Datum>): ReactElement {
    const {
        arcs,
        path,
        getKey,
        getColor,
    } = props

    return (
        <g>
            {
                arcs.map(arc =>
                    <PieArc
                        key={getKey(arc)}
                        arc={arc}
                        path={path}
                        getColor={getColor}
                        getKey={getKey}/>
                )
            }
        </g>
    );
}

interface PieArcProps<Datum> {
    getKey: (d: PieArcDatum<Datum>) => string;
    getColor: (d: PieArcDatum<Datum>) => string;
    path: ArcType<any, PieArcDatum<Datum>>;
    arc: PieArcDatum<Datum>;
}

function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { path, arc, getColor, getKey} = props;
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
        <g>
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
        </g>
    );
}
