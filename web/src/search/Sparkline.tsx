import { scaleLinear } from 'd3-scale'
import { select } from 'd3-selection'
import { line } from 'd3-shape'
import * as React from 'react'

interface Props {
    data: number[]
    width: number
    height: number
    isLightTheme: boolean
}

interface State {}

export class Sparkline extends React.PureComponent<Props, State> {
    private drawSparkline = (ref: SVGElement | null): void => {
        if (!ref) {
            return
        }

        const { data, width, height } = this.props
        if (!data.length) {
            return
        }
        const x = scaleLinear()
            .domain([-data.length, data.length])
            .range([0, width])

        const y = scaleLinear()
            .domain([-Math.max(...data), Math.max(...data)])
            .range([height, 0])
        const chartLine = line()
            .x((d, i) => x(i))
            .y((d, i) => y(Number(d)))

        const svg = select(ref!)
        svg.selectAll('*').remove()

        const strokeColor = this.props.isLightTheme ? '#cad2e2' : '#566e9f'

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
            .attr('height', height)
            .attr('overflow', 'visible')
        svg
            .append('circle')
            .attr('fill', strokeColor)
            .attr('cx', x(data.length - 1))
            .attr('cy', y(data[data.length - 1]))
            .attr('r', 2)
    }

    public render(): JSX.Element | null {
        const { width, height } = this.props

        return <svg viewBox={`0 0 200 20`} width={width} height={height} ref={this.drawSparkline} />
    }
}
