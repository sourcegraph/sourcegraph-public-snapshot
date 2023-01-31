export const defaultDurationValues = [
    // These values match the output of date-fns#formatDuration
    { value: 168, displayText: '7 days' },
    { value: 744, displayText: '1 month' }, // 31 days
    { value: 2160, displayText: '3 months' }, // 90 days
    { value: 8760, displayText: '1 year' }, // 365 days
    { value: 43824, displayText: '5 years' }, // 4 years + 1 leap day
]

export const formatDurationValue = (value: number): string => {
    const match = defaultDurationValues.find(candidate => candidate.value === value)
    if (!match) {
        return `${value} hours`
    }

    return match.displayText
}
