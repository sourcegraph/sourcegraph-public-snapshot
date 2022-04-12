import { MouseEvent } from 'react'

import { LineChartSeries } from 'sourcegraph'

export type YAccessor<Datum> = (data: Datum) => any

export interface LineChartSeriesWithData<Datum> extends LineChartSeries<Datum> {
    data: Point[]
}

export interface Point {
    x: Date | number
    y: number | null
}

/** Accessors map for getting values for x and y axes from datum object */
export interface Accessors<Datum, Key extends keyof Datum> {
    x: (d: Datum) => Date | number
    y: Record<Key, YAccessor<Datum>>
}

export interface DatumZoneClickEvent {
    originEvent: MouseEvent<unknown>
    link?: string
}

export type onDatumZoneClick = (event: DatumZoneClickEvent) => void
