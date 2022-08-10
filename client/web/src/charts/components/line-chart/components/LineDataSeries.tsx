import { ReactElement, SVGProps } from 'react'

import { Group } from '@visx/group'
import { LinePath } from '@visx/shape'
import { ScaleLinear, ScaleTime } from 'd3-scale'

import { Point } from '../types'
import { getDatumValue, isDatumWithValidNumber, SeriesDatum } from '../utils'

import { PointGlyph } from './PointGlyph'

const NULL_LINK = (): undefined => undefined

interface LineDataSeriesProps<D> extends SVGProps<SVGGElement> {
    id: string
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    dataset: SeriesDatum<D>[]
    color: string | undefined
    activePointId?: string
    getLinkURL?: (datum: D, index: number) => string | undefined

    onDatumClick: () => void
    onDatumFocus: (point: Point) => void
}

export function LineDataSeries<D>(props: LineDataSeriesProps<D>): ReactElement {
    const {
        id,
        xScale,
        yScale,
        dataset,
        color = 'green',
        activePointId,
        tabIndex,
        getLinkURL = NULL_LINK,
        onDatumClick,
        onDatumFocus,
        ...attributes
    } = props

    return (
        <Group tabIndex={tabIndex} {...attributes}>
            <LinePath
                data={dataset}
                defined={isDatumWithValidNumber}
                x={data => xScale(data.x)}
                y={data => yScale(getDatumValue(data))}
                stroke={color}
                strokeLinecap="round"
                strokeWidth={2}
            />

            {dataset.map((datum, index) => {
                const datumValue = getDatumValue(datum)
                const link = getLinkURL(datum.datum, index)
                const pointId = `${id}-${index}`

                return (
                    <PointGlyph
                        key={pointId}
                        tabIndex={tabIndex}
                        top={yScale(datumValue)}
                        left={xScale(datum.x)}
                        active={activePointId === pointId}
                        color={color}
                        linkURL={link}
                        onClick={onDatumClick}
                        onFocus={event =>
                            onDatumFocus({
                                id: pointId,
                                xValue: datum.x,
                                yValue: datumValue,
                                seriesId: id,
                                linkUrl: link,
                                node: event.target,
                            })
                        }
                    />
                )
            })}
        </Group>
    )
}
