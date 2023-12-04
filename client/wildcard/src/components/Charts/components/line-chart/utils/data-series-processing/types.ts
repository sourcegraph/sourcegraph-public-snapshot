import type { Series } from '../../../../types'
import { isValidNumber } from '../data-guards'

export enum SeriesType {
    Independent,
    Stacked,
}

export interface StackedSeries<Datum> extends Omit<Series<Datum>, 'data'> {
    type: SeriesType.Stacked
    data: StackedSeriesDatum<Datum>[]
}

export interface IndependentSeries<Datum> extends Omit<Series<Datum>, 'data'> {
    type: SeriesType.Independent
    data: StandardSeriesDatum<Datum>[]
}

export type SeriesWithData<Datum> = StackedSeries<Datum> | IndependentSeries<Datum>

export type SeriesDatum<Datum> = StandardSeriesDatum<Datum> | StackedSeriesDatum<Datum>

export interface StandardSeriesDatum<Datum> {
    id: string
    datum: Datum
    y: number | null
    x: Date
}

export interface StackedSeriesDatum<Datum> {
    id: string
    datum: Datum
    y1: number | null
    y0: number | null
    x: Date
}

export function isStandardSeriesDatum<Datum>(datum: SeriesDatum<Datum>): datum is StandardSeriesDatum<Datum> {
    return 'y' in datum && (isValidNumber(datum.y) || datum.y === null)
}
