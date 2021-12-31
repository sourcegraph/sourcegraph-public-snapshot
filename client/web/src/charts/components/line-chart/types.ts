export interface LineChartSeries<D> {
    /**
     * The key in each data object for the values this line should be
     * calculated from.
     */
    dataKey: keyof D

    /**
     * Links for data series points. Note that for points that don't have
     * values for the series this list still should have an undefined URL
     * in order to keep the linkURLs array and the common data array equal by length.
     */
    linkURLs?: string[]

    /** The name of the line shown in the legend and tooltip. */
    name?: string

    /** The CSS color of the line. */
    color?: string
}

export interface Point {
    id: string
    seriesKey: string
    index: number
    value: number
    color: string
    x: number
    y: number
    linkUrl?: string
}
