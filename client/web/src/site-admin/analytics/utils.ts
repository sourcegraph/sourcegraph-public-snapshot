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
                label: days === max ? `${days}+` : `${days}`,
                value: index >= 0 ? datums.slice(index).reduce((sum, datum) => sum + datum.frequency, 0) : 0,
            })
        } else if (index >= 0 && datums[index].daysUsed === days) {
            result.push({
                label: `${days}`,
                value: datums[index].frequency,
            })
        } else {
            result.push({
                label: `${days}+`,
                value: 0,
            })
        }
    }

    return result
}

export const formatNumber = (value: number): string => Intl.NumberFormat('en', { notation: 'compact' }).format(value)
