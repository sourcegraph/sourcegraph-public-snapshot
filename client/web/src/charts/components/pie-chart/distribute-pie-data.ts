/**
 * This algorithm sort origin data array in a way to alternate data for
 * better visual distribution of labels. When you have small parts of pie chart
 * together often you have label overlapping to avoid this we can change order
 * of data to => big arc => small arc => big arc. Just to add some space between labels.
 */
export function distributePieArcs<D>(data: readonly D[], getDatumValue: (datum: D) => number): D[] {
    const sortedData = [...data].sort((first, second) => getDatumValue(first) - getDatumValue(second))
    const result: D[] = []

    while (sortedData.length) {
        const firstElement = sortedData.shift()

        if (firstElement) {
            result.push(firstElement)
        }

        const lastElement = sortedData.pop()

        if (lastElement) {
            result.push(lastElement)
        }
    }

    return result
}
