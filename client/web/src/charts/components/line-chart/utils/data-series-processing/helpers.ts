import { isValidNumber } from '../data-guards'

import { isStandardSeriesDatum, SeriesDatum } from './types'

/**
 * Filters series data list, preserves null value at the beginning of the series data list
 * and removes null value between the points.
 *
 * ```
 * Null value ▽   Real point ■                  Null value ▽   Real point ■
 * ┌────────────────────────────────────┐       ┌────────────────────────────────────┐
 * │░░░░░░░░░░░░░░░                     │       │░░░░░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░                     │       │░░░░░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░                ■    │       │░░░░░░░░░░░░░░░                ■    │
 * │░░░░░░░░░░░░▽░░    ■                │       │░░░░░░░░░░░░▽░░    ■                │
 * │░░░░░░░░░░░░░░░          ▽          │──────▶│░░░░░░░░░░░░░░░                     │
 * │░░░░░░▽░░░░░░░░ ■                   │       │░░░░░░▽░░░░░░░░ ■                   │
 * │░░░░░░░░░░░░░░░       ■             │       │░░░░░░░░░░░░░░░       ■             │
 * │░░░▽░░░░░░░░░░░                     │       │░░░▽░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░             ▽       │       │░░░░░░░░░░░░░░░                     │
 * └────────────────────────────────────┘       └────────────────────────────────────┘
 *```
 */
export function getFilteredSeriesData<Datum>(data: SeriesDatum<Datum>[]): SeriesDatum<Datum>[] {
    const firstNonNullablePointIndex = Math.max(data.findIndex(isDatumWithValidNumber), 0)

    // Preserve null values at the beginning of the series data list
    // but remove null holes between the points further.
    const nullBeginningValues: SeriesDatum<Datum>[] = data.slice(0, firstNonNullablePointIndex)

    const pointsWithoutHoles = data
        // Get values after null area
        .slice(firstNonNullablePointIndex)
        .filter(isDatumWithValidNumber)

    return [...nullBeginningValues, ...pointsWithoutHoles]
}

export function isDatumWithValidNumber<Datum>(datum: SeriesDatum<Datum>): boolean {
    return isStandardSeriesDatum(datum) ? isValidNumber(datum.y) : isValidNumber(datum.y1)
}

export function getDatumValue<Datum>(datum: SeriesDatum<Datum>): number {
    return isStandardSeriesDatum(datum) ? datum.y ?? 0 : datum.y1 ?? 0
}
