import type { SearchBasedInsightSeries } from '../../../core'

export function getSanitizedLine(line: SearchBasedInsightSeries): SearchBasedInsightSeries {
    return {
        id: line.id,
        name: line.name.trim(),
        stroke: line.stroke,
        // Query field is a regexp field for code insight query setting
        // Native html input element adds escape symbols by itself
        // to prevent this behavior below we replace double escaping
        // with just one series of escape characters e.g. - //
        query: line.query.replaceAll('\\\\', '\\'),
    }
}

export function getSanitizedSeries(rawSeries: SearchBasedInsightSeries[]): SearchBasedInsightSeries[] {
    return rawSeries.map(getSanitizedLine)
}
