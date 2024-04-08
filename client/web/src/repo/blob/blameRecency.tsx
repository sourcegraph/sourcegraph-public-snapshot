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
 * The "direction" of color can be reversed, so that the most recent commits are light
 * and the oldest are dark (e.g. for dark mode).
 *
 * @param commitDate The date of the commit
 * @param lightToDark If true, the most recent commits will be light and the oldest dark. Default is false.
 * @returns The color for the recency of the commit. It's an `rgb(...)` CSS color string.
 */
export function getBlameRecencyColor(commitDate: Date | undefined, lightToDark = false): string {
    if (!commitDate) {
        return 'var(--gray-04)'
    }

    const age = (Date.now() - commitDate.getTime()) / MILLIS_IN_HOUR

    // Get a value between [0, 1) that represents the recency of the commit
    // (0 is most recent, 1 is least recent)
    let recency = age / (age + MIDPOINT)

    // The color scheme goes from light (0) to dark (1), but unless lightToDark is true,
    // we want the most recent commits to be dark and the oldest to be light. Therefore
    // we simply invert the recency value.
    if (!lightToDark) {
        recency = 1 - recency
    }

    return interpolatePurples(recency)
}
