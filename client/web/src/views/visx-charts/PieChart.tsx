import React from 'react';
import Pie, { ProvidedProps, PieArcDatum } from '@visx/shape/lib/shapes/Pie';
import { scaleOrdinal } from '@visx/scale';
import { Group } from '@visx/group';
// TODO Replace import below on standard visx/annotation package
import { Annotation, Connector, Label } from './annotation';
import browserUsage, { BrowserUsage as Browsers } from '@visx/mock-data/lib/mocks/browserUsage';
import { Arc as ArcType } from 'd3-shape';

// data and types
type BrowserNames = keyof Browsers;

interface BrowserUsage {
    label: BrowserNames;
    usage: number;
}

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
const browserNames = Object.keys(browserUsage[0]).filter(k => k !== 'date' && browserUsage[0][k] > 5) as BrowserNames[];
const browsers: BrowserUsage[] = browserNames.map(name => ({
    label: name,
    usage: Number(browserUsage[0][name]),
}));

// accessor functions
const usage = (data: BrowserUsage): number => data.usage;

// color scales
const getBrowserColor = scaleOrdinal({
    domain: browserNames,
    range: [
        'red',
        'green',
        'pink',
        'wheat',
        'darkgreen',
    ],
});

const defaultMargin = { top: 20, right: 20, bottom: 20, left: 20 };

export type PieProps = {
    width: number;
    height: number;
    margin?: typeof defaultMargin;
};

export function PieExample(props: PieProps) {
    const {
        width,
        height,
        margin = defaultMargin,
    } = props;

    if (width < 10) return null;

    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const radius = Math.min(innerWidth, innerHeight) / 3;
    const centerY = innerHeight / 2;
    const centerX = innerWidth / 2;
    // const donutThickness = 50;

    return (
        <svg width={width} height={height}>
            <Group top={centerY + margin.top} left={centerX + margin.left}>
                <Pie
                    data={browsers}
                    pieValue={usage}
                    outerRadius={radius}
                    cornerRadius={3}
                    padRadius={30}
                >
                    {pie => (
                        <PieArcs<BrowserUsage>
                            {...pie}
                            getKey={arc => arc.data.label}
                            getColor={arc => getBrowserColor(arc.data.label)}
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

function PieArcs<Datum>(props: AnimatedPieProps<Datum>) {
    const {
        arcs,
        path,
        getKey,
        getColor,
    } = props

    return (
        <g>
            { arcs.map(arc =>
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

function PieArc<Datum>(props: PieArcProps<Datum>) {
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
                stroke={'white'}
                strokeWidth={1}
            />
            <Annotation
                x={surfaceX}
                y={surfaceY}
                dx={labelX}
                dy={labelY}
            >
                <Connector type={"line"} />
                <Label
                    showBackground={false}
                    showAnchorLine={false}
                    title={getKey(arc)}
                    subtitle={`${arc.value}%`} />
            </Annotation>

            <circle r={4} fill={'black'} cx={surfaceX + labelX} cy={surfaceY + labelY}/>
        </g>
    );
}
