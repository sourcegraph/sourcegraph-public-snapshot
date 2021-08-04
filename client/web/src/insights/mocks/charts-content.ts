import { LineChartContent } from 'sourcegraph'

export const LINE_CHART_CONTENT_MOCK: LineChartContent<any, string> = {
    chart: 'line',
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 0, b: 150 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 20, b: 260 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 56, b: 200 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 98, b: 190 },
        { x: 1588965700286, a: 123, b: 170 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            stroke: 'var(--warning)',
            linkURLs: [
                '#A:1st_data_point',
                '#A:2nd_data_point',
                '#A:3rd_data_point',
                '#A:4th_data_point',
                '#A:5th_data_point',
            ],
        },
        {
            dataKey: 'b',
            name: 'B metric',
            stroke: 'var(--warning)',
            linkURLs: [
                '#B:1st_data_point',
                '#B:2nd_data_point',
                '#B:3rd_data_point',
                '#B:4th_data_point',
                '#B:5th_data_point',
            ],
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time',
        type: 'number',
    },
}
