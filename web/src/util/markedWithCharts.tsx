import { extent, max } from 'd3-array'
import { axisBottom, axisLeft } from 'd3-axis'
import { scaleBand, scaleLinear } from 'd3-scale'
import { select } from 'd3-selection'
import { curveBasis, line as d3line } from 'd3-shape'
import yaml from 'js-yaml'
import marked from 'marked'
import { Props as MarkvisProps } from 'markvis'
import bar from 'markvis-bar'
import line from 'markvis-line'

declare type charts = 'line' | 'bar'

const layouts: { [T in charts]: (props: any) => string } = {
    line,
    bar,
}

export function markedWithCharts(
    src: string,
    options?: marked.MarkedOptions | undefined,
    callback?: ((error: any, parseResult: string) => void) | undefined
): string {
    const renderer = new marked.Renderer()
    renderer.code = (code: string, language: string, escaped: boolean): string => {
        if (language === 'vis') {
            let userOpts: MarkvisProps
            try {
                userOpts = yaml.safeLoad(code)
            } catch (err) {
                console.error('error parsing markvis options, must be valid YAML', err)
                throw err
            }
            const chart: charts = (userOpts.layout as charts) || 'bar'
            const chartID = Math.random()
                .toString(36)
                .substr(2, 5)
            const opts = Object.assign(
                // Assign Sourcegraph default chart options.
                {
                    isCurve: false,
                    margin: { top: 20, right: 20, left: 50, bottom: 50 },
                    showYAxis: false,
                    showValues: true,
                    showDots: true,
                    width: 800,
                    height: 250,
                    barColor: '#566e9f',
                    barHoverColor: '#a2b0cd',
                    lineColor: '#566e9f',
                },
                userOpts,
                // Options that can't be overwritten.
                {
                    d3: {
                        curveBasis,
                        extent,
                        line: d3line,
                        select,
                        scaleBand,
                        scaleLinear,
                        axisBottom,
                        axisLeft,
                        max,
                    },
                    container: `<div class="markvis-chart-container"><div class="markvis-chart" id="markvis-chart-${chartID}"></div></div>`,
                    selector: `#markvis-chart-${chartID}`,
                    barAttrs: { 'data-tooltip': (v: number) => v },
                    dotAttrs: { 'data-tooltip': (v: number) => v },
                }
            )
            return layouts[chart](opts)
        }
        return renderer.code(code, language, escaped)
    }
    return marked(src, Object.assign({}, options, { renderer }), callback)
}
