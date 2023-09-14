import { isValidNumber } from '../data-guards'

import { isStandardSeriesDatum, type SeriesDatum } from './types'

export function isDatumWithValidNumber<Datum>(datum: SeriesDatum<Datum>): boolean {
    return isStandardSeriesDatum(datum) ? isValidNumber(datum.y) : isValidNumber(datum.y1)
}

export function getDatumValue<Datum>(datum: SeriesDatum<Datum>): number {
    return isStandardSeriesDatum(datum) ? datum.y ?? 0 : datum.y1 ?? 0
}

export function encodePointId(seriesId: string | number, pointIndex: number): string {
    // Encode all special symbols that may be in the original series id such as
    // '"' symbol that breaks element query selectors in the chart where we use this id
    // for chart points and lines.
    // See https://github.com/sourcegraph/sourcegraph/issues/45376
    return encodeURIComponent(`${pointIndex}.${seriesId.toString()}`)
}

export function decodePointId(id: string): [string, number] {
    const [index, ...seriesIdParts] = decodeURIComponent(id).split('.')
    return [seriesIdParts.join('.'), +index]
}
