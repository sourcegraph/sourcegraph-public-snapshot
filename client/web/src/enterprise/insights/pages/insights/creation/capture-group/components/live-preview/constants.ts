import { SeriesChartContent } from '../../../../../../core/backend/code-insights-backend-types'

interface MockDatum {
    x: number
    a: number
    b: number
}

export const MOCK_CHART_CONTENT: SeriesChartContent<MockDatum> = {
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
            name: 'Go 1.11',
            color: 'var(--oc-indigo-7)',
        },
        {
            dataKey: 'b',
            name: 'Go 1.12',
            color: 'var(--oc-orange-7)',
        },
    ],
    getXValue: datum => new Date(datum.x),
}
