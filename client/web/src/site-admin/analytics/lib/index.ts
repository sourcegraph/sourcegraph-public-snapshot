export const formatNumber = (value: number): string => Intl.NumberFormat('en', { notation: 'compact' }).format(value)
