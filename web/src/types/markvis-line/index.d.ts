declare module 'markvis-line' {
    import d3 = __d3

    export interface Props {
        /**
         * Data from file or web processed by d3 library.
         */
        data: any[]

        /**
         * [d3](https://github.com/d3/d3) library which used in browser environment.
         */
        d3?: d3

        /**
         * [d3-node](https://github.com/d3-node/d3-node) constructor which used in node environment.
         */
        d3node?: any

        /**
         * DOM selector in container.
         * @default '#chart'
         */
        selector?: string

        /**
         * DOM contained the visualization result.
         * @default '<div id="container"><h2>Bar Chart</h2><div id="chart"></div></div>'
         */
        container?: string

        /**
         * Bar chart style.
         * @default
         *   .bar {fill: steelblue;}
         *   .bar:hover {fill: brown;}
         */
        style?: string

        /**
         * Line dot element attributes.
         * @default {}
         */
        dotAttrs?: { [key: string]: (x: any) => string }

        /**
         * SVG width for bar chart.
         * @default 960
         */
        width?: number

        /**
         * SVG height for bar chart.
         * @default 500
         */
        height?: number

        /**
         * Whether the chart should be automatically resized to fit its container.
         * If true, the `width` and `height` options are used for the initial sizing/SVG viewBox size.
         * @default true
         */
        responsive?: boolean

        /**
         * Margin of the first wrapper in SVG, usually used to add axis.
         * @default { top: 20, right: 20, bottom: 20, left: 20 }
         */
        margin?: { top: number; right: number; bottom: number; left: number }

        /**
         * Color of line.
         * @default steelblue
         */
        lineColor?: string

        /**
         * Width of line.
         * @default 1.5
         */
        lineWidth?: number

        /**
         * Whether to render the X axis or not.
         * @default true
         */
        showXAxis?: boolean

        /**
         * Whether to render the Y axis or not.
         * @default true
         */
        showYAxis?: boolean

        /**
         * Whether to render bar value labels or not above each bar.
         * @default true
         */
        showValues?: boolean

        /**
         * Whether the line is a curve.
         * @default true
         */
        isCurve?: boolean

        /**
         * Whether to export to a PNG image.
         * @default false
         */
        export?: boolean
    }

    export default function line(props: Props): string
}
