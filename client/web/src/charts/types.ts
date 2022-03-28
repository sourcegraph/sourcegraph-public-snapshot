export enum SeriesBasedChartTypes {
    Line,
}

export enum CategoricalBasedChartTypes {
    Pie,
    Bar,
}

export interface Series<Datum> {
    /**
     * The key in each data object for the values this line should be
     * calculated from.
     */
    dataKey: keyof Datum

    /**
     * Link for data series point. It may be used to make datum points with links
     * instead of plain visual svg elements.
     */
    getLinkURL?: (datum: Datum) => string | undefined

    /**
     * The name of the line shown in the legend and tooltip
     */
    name: string

    /**
     * The CSS color of the series. If color wasn't provided the default (gray) color
     * will be used instead.
     */
    color?: string
}
