import { isValidNumber } from '../data-guards'

import { isStandardSeriesDatum, SeriesDatum } from './types'

export function isDatumWithValidNumber<Datum>(datum: SeriesDatum<Datum>): boolean {
    return isStandardSeriesDatum(datum) ? isValidNumber(datum.y) : isValidNumber(datum.y1)
}

export function getDatumValue<Datum>(datum: SeriesDatum<Datum>): number {
    return isStandardSeriesDatum(datum) ? datum.y ?? 0 : datum.y1 ?? 0
}
