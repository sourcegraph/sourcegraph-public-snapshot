import { Optional } from 'utility-types'

interface GetChartContentSizesInput {
    width: number
    height: number
    margin?: Optional<Margin>
}

interface Margin {
    top: number
    right: number
    bottom: number
    left: number
}

interface ChartContentSizes {
    width: number
    height: number
    margin: Margin
}

export function getChartContentSizes(input: GetChartContentSizesInput): ChartContentSizes {
    const { width, height, margin = {} } = input

    const { top, left, bottom, right } = {
        top: margin.top ?? 0,
        right: margin.right ?? 0,
        bottom: margin.bottom ?? 0,
        left: margin.left ?? 0,
    }
    return {
        width: width - left - right,
        height: height - top - bottom,
        margin: { top, left, bottom, right },
    }
}
