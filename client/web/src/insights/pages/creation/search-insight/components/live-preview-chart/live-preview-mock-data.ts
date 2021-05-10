import type { LineChartContent } from 'sourcegraph';

export const DEFAULT_MOCK_CHART_CONTENT: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 110, b: 150 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 145, b: 260 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: 200 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 134, b: 190 },
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
        scale: 'time' as const,
        type: 'number',
    },
}
