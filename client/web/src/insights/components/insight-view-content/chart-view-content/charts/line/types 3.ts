import { MouseEvent } from 'react'

export type YAccessor<Datum> = (data: Datum) => any

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
