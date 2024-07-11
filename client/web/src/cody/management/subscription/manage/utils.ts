// usdCentsToHumanString converts an amount into a human-friendly display.
// e.g. 42 -> "$0.42", 2048 -> "$20.48", 123456 -> "$1,234.56".
export function usdCentsToHumanString(usdCents: number): string {
    const commaSeparatedDollars = Math.floor(usdCents / 100).toLocaleString('en-US')
    const paddedCentsStr = `${usdCents % 100}`.padStart(2, '0')
    return `$${commaSeparatedDollars}${paddedCentsStr !== '00' ? `.${paddedCentsStr}` : ''}`
}

// humanizeDate returns a friendly display for an ISO-8601 time.
// "2024-01-10T13:56:24-08:00" -> "January 10th, 2024"
export function humanizeDate(dateString: string): string {
    const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'long', day: 'numeric' }
    return new Intl.DateTimeFormat('en-US', options).format(new Date(dateString))
}
