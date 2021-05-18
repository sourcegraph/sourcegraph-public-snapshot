import { random } from 'lodash'
import type { LineChartContent } from 'sourcegraph'

export const DEFAULT_MOCK_CHART_CONTENT: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, a: 20, b: 200 },
        { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, a: 40, b: 177 },
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 110, b: 150 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 105, b: 165 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 160, b: 100 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 184, b: 85 },
        { x: 1588965700286, a: 200, b: 50 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'Old gql imports',
            stroke: 'var(--oc-indigo-7)',
        },
        {
            dataKey: 'b',
            name: 'New gql operation imports',
            stroke: 'var(--oc-orange-7)',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

export const getRandomDataForMock = (): unknown[] =>
    new Array(6).fill(null).map((item, index) => ({
        x: 1588965700286 - index * 24 * 60 * 60 * 1000,
        a: random(20, 200),
        b: random(10, 200),
    }))
