import { scaleLinear } from 'd3-scale'
import { select } from 'd3-selection'
import { line } from 'd3-shape'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { colorTheme, getColorTheme } from '../settings/theme'

interface Props {
    data: number[]
    width: number
    height: number
}

interface State {
    isLightTheme: boolean
}

export class Sparkline extends React.Component<Props, State> {
    private svgRef?: SVGElement | null
    private subscriptions = new Subscription()

    public state: State = { isLightTheme: getColorTheme() === 'light' }

    public componentDidMount(): void {
        this.subscriptions.add(
            colorTheme.subscribe(theme => {
                this.setState({ isLightTheme: theme === 'light' }, () => this.drawSparkline())
            })
        )
        this.drawSparkline()
    }

    public shouldComponentUpdate(): boolean {
        this.drawSparkline()
        return true
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private drawSparkline(): void {
        const { data, height } = this.props
        if (!data.length) {
            return
        }
        const width = this.props.width - 5
        const x = scaleLinear()
            .domain([0, data.length])
            .range([0, width])

        const y = scaleLinear()
            .domain([Math.min(...data), Math.max(...data)])
            .range([height * 0.9, 0])
        const chartLine = line()
            .x((d, i) => x(i))
            .y((d, i) => y(Number(d)))

        const svg = select(this.svgRef!)
        svg.selectAll('*').remove()

        const { isLightTheme } = this.state
        const strokeColor = isLightTheme ? '#cad2e2' : '#566e9f'

        svg
            .append('path')
            .datum(data)
            .attr('fill', 'none')
            .attr('stroke', strokeColor)
            .attr('stroke-linejoin', 'round')
            .attr('stroke-linecap', 'round')
            .attr('stroke-width', 2)
            .attr('d', chartLine(data as any) as any)
            .attr('width', width)
            .attr('height', height * 0.8)
            .attr('overflow', 'visible')
        svg
            .append('circle')
            .attr('fill', strokeColor)
            .attr('cx', x(data.length - 1))
            .attr('cy', y(data[data.length - 1]))
            .attr('r', 4)
    }

    public render(): JSX.Element | null {
        const { width, height } = this.props

        return <svg width={width} height={height} ref={ref => (this.svgRef = ref)} />
    }
}
