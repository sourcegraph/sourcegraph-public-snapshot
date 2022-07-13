import { addDays, getDayOfYear, startOfDay, startOfWeek, sub } from 'date-fns'

import { AnalyticsDateRange } from '../../graphql-operations'

export interface FrequencyDatum {
    label: string
    value: number
}

export interface StandardDatum {
    date: Date
    value: number
}

export function buildFrequencyDatum(
    datums: { daysUsed: number; frequency: number }[],
    min: number,
    max: number,
    isGradual = true
): FrequencyDatum[] {
    const result: FrequencyDatum[] = []
    for (let days = min; days <= max; ++days) {
        const index = datums.findIndex(datum => datum.daysUsed >= days)
        if (isGradual || days === max) {
            result.push({
                label: `${days} days`,
                value: index >= 0 ? datums.slice(index).reduce((sum, datum) => sum + datum.frequency, 0) : 0,
            })
        } else if (index >= 0 && datums[index].daysUsed === days) {
            result.push({
                label: `${days} days`,
                value: datums[index].frequency,
            })
        } else {
            result.push({
                label: `${days}+ days`,
                value: 0,
            })
        }
    }

    return result
}

export function buildStandardDatum(datums: StandardDatum[], dateRange: AnalyticsDateRange): StandardDatum[] {
    // Generates 0 value series for dates that don't exist in the original data
    const [to, daysOffset] =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? [startOfWeek(new Date(), { weekStartsOn: 1 }), 7]
            : [startOfDay(new Date()), 1]
    const from =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? sub(to, { months: 3 })
            : dateRange === AnalyticsDateRange.LAST_MONTH
            ? sub(to, { months: 1 })
            : sub(to, { weeks: 1 })
    const newDatums: StandardDatum[] = []
    let date = to
    while (date >= from) {
        const datum = datums?.find(datum => getDayOfYear(datum.date) === getDayOfYear(date))
        newDatums.push(datum ? { ...datum, date } : { date, value: 0 })
        date = addDays(date, -daysOffset)
    }

    return newDatums
}

export const formatNumber = (value: number): string => Intl.NumberFormat('en', { notation: 'compact' }).format(value)
