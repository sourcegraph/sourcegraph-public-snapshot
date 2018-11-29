declare module 'markvis' {
    export interface Props {
        /**
         * Data from file or web processed by d3 library.
         */
        data: any[]

        /**
         * [d3](https://github.com/d3/d3) library which used in browser environment.
         */
        d3?: any

        /**
         * [d3-node](https://github.com/d3-node/d3-node) constructor which used in node environment.
         */
        d3node?: any

        /**
         * Name of chart layout. You can customize any chart layout you want.
         */
        layout: string

        /**
         * Customized renderer to render a new layout you want.
         */
        // render: (any) => any
        // chart: (any) => any

        /**
         * DOM contained the visualization result.
         * @default: '<div id="container"><h2>Bar Chart</h2><div id="chart"></div></div>'
         */
        container: string

        /**
         * DOM selector in container.
         * @default: '#chart'
         */
        selector: string

        /**
         * Bar chart style.
         * @default: ''
         */
        style: string

        /**
         * SVG width for bar chart.
         * @default: 960
         */
        width: number

        /**
         * SVG height for bar chart.
         * @default: 500
         */
        height: number

        /**
         * Margin of the first wrapper in SVG, usually used to add axis.
         * @default: { top: 20, right: 20, bottom: 20, left: 20 }
         */
        margin: { top: number; right: number; bottom: number; left: number }
    }

    export default function render(options: any): (tokens: any, idx: any, _options: any, env: any) => string
}
