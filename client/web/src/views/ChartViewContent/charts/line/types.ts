import { MouseEvent } from 'react'

/** Accessors map for getting values for x and y axes from datum object */
export interface Accessors<Datum, Key extends keyof Datum> {
    x: (d: Datum) => Date | number
    y: Record<Key, (data: Datum) => any>
}

export interface DatumZoneClickEvent {
    originEvent: MouseEvent<unknown>
    link?: string
}

export type onDatumZoneClick = (event: DatumZoneClickEvent) => void
