export interface FrequencyDatum {
    label: string
    value: number
}

export interface StandardDatum {
    date: Date
    value: number
}

export function buildFrequencyDatum(
    datums: { daysUsed: number; frequency: number; percentage: number }[],
    uniqueOrPercentage: 'unique' | 'percentage',
    max: number
): FrequencyDatum[] {
    const result: FrequencyDatum[] = []
    // loop from 30+ days to -> 1 day
    for (let index = max - 1; index >= 0; index--) {
        const daysUsed = index + 1
        const datum = datums.find(data => data.daysUsed === daysUsed)

        if (datum) {
            result.push({
                label: `${daysUsed}`,
                value: Math.round(datum[uniqueOrPercentage === 'unique' ? 'frequency' : 'percentage'] * 100) / 100, // round to .2,
            })
        } else {
            result.push({
                label: daysUsed === max ? `${daysUsed}+` : `${daysUsed}`,
                // if no item for 18 days in datums then copy value from last result item, i.e. 19 days.
                value: result[result.length - 1]?.value || 0,
            })
        }
    }

    return result
}

export const formatNumber = (value: number): string => Intl.NumberFormat('en', { notation: 'compact' }).format(value)
