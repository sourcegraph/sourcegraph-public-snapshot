export const defaultDurationValues = [
    { value: null, displayText: 'Forever' },
    { value: 168, displayText: '1 week' }, // 168 hours
    { value: 672, displayText: '1 month' }, // 168 hours * 4
    { value: 2016, displayText: '3 months' }, // 168 hours * 4 * 3
    { value: 4032, displayText: '6 months' }, // 168 hours * 4 * 6
    { value: 8064, displayText: '1 year' }, // 168 hours * 4 * 12
    { value: 40320, displayText: '5 years' }, // 168 hours * 4 * 12 * 5
]

export const formatDurationValue = (value: number): string => {
    const match = defaultDurationValues.find(candidate => candidate.value === value)
    if (!match) {
        return `${value} hours`
    }

    return match.displayText
}
