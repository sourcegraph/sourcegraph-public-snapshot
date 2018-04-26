import { axisBottom, AxisContainerElement } from 'd3-axis'
import { scaleBand, scaleLinear, scaleOrdinal } from 'd3-scale'
import { select, Selection } from 'd3-selection'
import { stack } from 'd3-shape'
import { isEqual } from 'lodash'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'

interface BarChartSeries {
    [key: string]: null
}

interface BarChartDatum<T extends BarChartSeries> {
    xLabel: string
    yValues: { [key in keyof T]: number }
}

interface Props<T extends BarChartSeries> {
    /**
     * Bar chart data.
     * One datum for each column, with each datum containing values for each series in the given column.
     */
    data: BarChartDatum<T>[]
    /**
     * Initial width (chart will be automatically resized to fit container).
     */
    width: number
    /**
     * Initial height (chart will be automatically resized to fit container).
     */
    height: number
    /**
     * Display column totals labels.
     */
    showLabels?: boolean
    /**
     * Display legend.
     */
    showLegend?: boolean
    className?: string
    isLightTheme: boolean
}

interface State<T extends BarChartSeries> {
    /**
     * Bar chart data.
     * One datum for each column, with each datum containing values for each series in the given column.
     */
    data: BarChartDatum<T>[]
    /**
     * Bar colors for each series.
     */
    barColors: string[]
    /**
     * Series names.
     */
    series: string[]
    /**
     * x-axis labels for each column.
     */
    xLabels: string[]
    /**
     * Values for each series in each column.
     */
    yValues: { [key in keyof T]: number }[]
    /**
     * Total heights of each column.
     */
    yHeights: number[]
}

export class BarChart<T extends BarChartSeries> extends React.PureComponent<Props<T>, State<T>> {
    private propsChanges = new Subject<Props<T>>()
    private subscriptions = new Subscription()

    private svgRef: SVGSVGElement | null = null

    public constructor(props: Props<T>) {
        super(props)
        this.state = {
            data: [],
            barColors: [],
            series: [],
            xLabels: [],
            yValues: [],
            yHeights: [],
        }
        this.subscriptions.add(
            this.propsChanges
                .pipe(distinctUntilChanged((a, b) => isEqual(a, b)))
                .subscribe(props => this.refreshChartDataAndRedraw(props))
        )
    }

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
    }

    public componentDidUpdate(): void {
        this.propsChanges.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private refreshChartDataAndRedraw(props: Props<T>): void {
        const data = props.data.reverse()
        this.setState(
            {
                data,
                barColors: props.isLightTheme ? ['#a2b0cd', '#cad2e2'] : ['#566e9f', '#a2b0cd'],
                series: Object.keys(data[0].yValues),
                xLabels: data.map(({ xLabel }) => xLabel),
                yValues: data.map(({ yValues }) => yValues),
                yHeights: data.map(({ yValues }) => Object.keys(yValues).reduce((acc, k) => acc + yValues[k], 0)),
            },
            () => {
                if (this.svgRef) {
                    this.drawChart(this.svgRef)
                }
            }
        )
    }

    private drawChart = (ref: SVGElement | null): void => {
        if (!ref) {
            return
        }
        const { width, height } = this.props
        if (!this.state.data.length) {
            return
        }

        const columns = this.state.xLabels.length

        const x = scaleBand()
            .domain(this.state.xLabels)
            .rangeRound([0, width])
        const y = scaleLinear()
            .domain([0, Math.max(...this.state.yHeights)])
            .range([height, 0])
        const z = scaleOrdinal<string, string>()
            .domain(this.state.series)
            .range(this.state.barColors)
        const xAxis = axisBottom(x)

        const svg = select(ref!)
        svg.selectAll('*').remove()

        const barWidth = width / columns - 2

        const barHolder = svg
            .classed(`d3-bar-chart ${this.props.className || ''}`, true)
            .attr('preserveAspectRatio', 'xMinYMin')
            .append('g')
            .classed('bar-holder', true)

        const stackData = stack()
            .keys(this.state.series)
            .value((d, key) => d[key])(this.state.yValues, this.state.series)

        // Generate bars.
        barHolder
            .append('g')
            .selectAll('g')
            .data(stackData)
            .enter()
            .append('g')
            .attr('fill', d => z(d.key))
            .selectAll('rect')
            .data(d => d)
            .enter()
            .append('rect')
            .classed('bar', true)
            .attr('x', (d, i) => x(this.state.xLabels[i]) || 0 + 1)
            .attr('y', d => y(d[1]))
            .attr('width', barWidth)
            .attr('height', d => y(d[0]) - y(d[1]))
            .attr('data-tooltip', (d, i) => `${d[1] - d[0]} users`)

        if (this.props.showLabels) {
            // Generate value labels on top of each column.
            barHolder
                .append('g')
                .selectAll('text')
                .data(this.state.data)
                .enter()
                .append('text')
                .attr('text-anchor', 'middle')
                .attr('x', d => x(d.xLabel) || 0)
                .attr('dx', barWidth / 2)
                .attr('y', (d, i) => y(this.state.yHeights[i]))
                .attr('dy', '-0.5em')
                .text((d, i) => this.state.yHeights[i])
        }

        // Generate x-axis and labels.
        barHolder
            .append<AxisContainerElement>('g')
            .classed('axis', true)
            .attr('transform', 'translate(0,' + height + ')')
            .call(xAxis)
            .selectAll('.tick text')
            .call(wrapLabel, barWidth)

        if (this.props.showLegend) {
            // Generate a legend.
            const legend = barHolder
                .append('svg')
                .attr('y', '-5em')
                .append('g')
                .attr('text-anchor', 'end')
                .selectAll('g')
                .data(this.state.series.slice().reverse())
                .enter()
                .append('g')
                .attr('transform', (d, i) => 'translate(0,' + i * 20 + ')')
            legend
                .append('rect')
                .attr('x', width - 19)
                .attr('width', 19)
                .attr('height', 19)
                .attr('fill', z)
            legend
                .append('text')
                .attr('x', width - 24)
                .attr('y', 9.5)
                .attr('dy', '0.32em')
                .text(d => d)
        }
    }

    public render(): JSX.Element | null {
        const { width, height } = this.props
        return <svg viewBox={`0 0 ${width} ${height}`} ref={ref => (this.svgRef = ref)} />
    }
}

// Source: Mike Bostock's "Wrapping Long Labels": https://bl.ocks.org/mbostock/7555321
function wrapLabel(text: Selection<any, any, any, any>, width: number): void {
    text.each(function(): void {
        const text = select(this)
        const words = text
            .text()
            .split(/\s+/)
            .reverse()
        const lineHeight = 1.1
        const y = text.attr('y')
        const dy = parseFloat(text.attr('dy'))
        let lineNumber = 0
        let currentWord
        // currentLine holds the line as it grows, until it overflows.
        let currentLine: string[] = []
        // tspan holds the current <tspan> element as it grows, until it overflows.
        let tspan = text
            .text(null)
            .append('tspan')
            .attr('x', 0)
            .attr('y', y)
            .attr('dy', dy + 'em')

        while (words.length) {
            currentWord = words.pop() || ''
            currentLine.push(currentWord)
            tspan.text(currentLine.join(' '))
            if ((tspan.node() as SVGTextContentElement).getComputedTextLength() > width) {
                currentLine.pop()
                tspan.text(currentLine.join(' '))
                // Start a new line and generate a new tspan element.
                currentLine = [currentWord]
                tspan = text
                    .append('tspan')
                    .attr('x', 0)
                    .attr('y', y)
                    .attr('dy', ++lineNumber * lineHeight + dy + 'em')
                    .text(currentWord)
            }
        }
    })
}
