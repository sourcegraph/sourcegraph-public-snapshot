import { LineChartContent } from 'sourcegraph'

export const LINE_CHART_CONTENT_MOCK: LineChartContent<any, string> = {
    chart: 'line',
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 4000, b: 15000 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 4000, b: 26000 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 5600, b: 20000 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 9800, b: 19000 },
        { x: 1588965700286, a: 12300, b: 17000 },
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

export const LINE_CHART_CONTENT_MOCK_EMPTY: LineChartContent<any, string> = {
    chart: 'line',
    data: [],
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            stroke: 'var(--warning)',
            linkURLs: [],
        },
        {
            dataKey: 'b',
            name: 'B metric',
            stroke: 'var(--warning)',
            linkURLs: [],
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time',
        type: 'number',
    },
}
