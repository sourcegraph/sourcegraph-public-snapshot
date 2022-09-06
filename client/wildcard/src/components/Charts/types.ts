import React from 'react'

export interface SeriesLikeChart<Datum> {
    series: Series<Datum>[]
    stacked?: boolean

    /**
     * Callback runs whenever a point-zone (zone around point) and point itself
     * on the chart is clicked.
     */
    onDatumClick?: (event: React.MouseEvent) => void
}

export interface CategoricalLikeChart<Datum> {
    data: Datum[]
    getDatumValue: (datum: Datum) => number
    getDatumName: (datum: Datum) => string
    getDatumHover?: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
    getDatumLink?: (datum: Datum) => string | undefined
    onDatumLinkClick?: (event: React.MouseEvent, datum: Datum, index: number) => void
}

export interface Series<Datum> {
    /** Unique series id. */
    id: string | number

    /** The name of the line shown in the legend and tooltip */
    name: string

    /*
     * List of datum (points) for this particular data series. Should
     * contain y and x-axis value for the point.
     */
    data: Datum[]

    /**
     * Getters that will run over each data points and should return x and y
     * value for series points.
     */
    getXValue: (datum: Datum) => Date
    getYValue: (datum: Datum) => number

    /**
     * Link for data series point. It may be used to make datum points with links
     * instead of plain visual svg elements.
     */
    getLinkURL?: (datum: Datum, index: number) => string | undefined

    /**
     * The CSS color of the series. If color wasn't provided the default (gray) color
     * will be used instead.
     */
    color?: string
}
