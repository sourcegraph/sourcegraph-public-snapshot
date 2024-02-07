import { interpolatePurples } from 'd3-scale-chromatic'

// MIDPOINT is the duration for which the scale function returns the midpoint color (0.5)
// The closer the midpoint is to "now", the more pronounced the color difference for
// recent commits. I.e. more recent commits will be easier to distinguish from each other
// than older commits.
// Conversely, if the midpoint is further back in time, the color difference for recent
// commits is less pronounced.
// Another factor is the granularity of time. If the scale is in hours, the color difference
// for commits from the last few days is more pronounced than if the scale is in months.

const MIDPOINT = 6 * 30 * 24 // 6 months in hours
const MILLIS_IN_HOUR = 1000 * 60 * 60

/**
 * Get the color for the recency of a commit. The color is interpolated between
 * light grey and purple, with grey being the oldest and purple being the most recent.
 * Dark to light is the default, but can be reversed (e.g. for dark mode).
 */
export function getBlameRecencyColor(commitDate: Date | undefined, darkToLight = true): string {
    if (!commitDate) {
        return 'var(--gray-04)'
    }

    const age = (Date.now() - commitDate.getTime()) / MILLIS_IN_HOUR

    // Get a value between [1, 0) that represents the recency of the commit
    // (1 is most recent, 0 is least recent)
    let recency = age / (age + MIDPOINT)
    if (darkToLight) {
        recency = 1 - recency
    }

    return interpolatePurples(recency)
}
