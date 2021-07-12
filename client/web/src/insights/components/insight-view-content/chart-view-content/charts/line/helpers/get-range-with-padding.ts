/**
 * Generate a new domain range for d3 scale configuration,
 * See https://github.com/d3/d3-scale#continuous_domain
 * Generally we can't add visual padding for y axis,
 * but we can calculate synthetic min and max values according padding value
 *
 * @param range - origin min-max range
 * @param paddingCoefficient number from 0 to 1 stands for how much padding
 * do we want to add in percentage terms
 */
export function getRangeWithPadding(range: [number, number], paddingCoefficient: number): [number, number] {
    const [min, max] = range
    const increment = ((max - min) * paddingCoefficient) / 2

    // Minimal value for insight line chart can't be less than 0
    return [Math.max(min - increment, 0), max + increment]
}
