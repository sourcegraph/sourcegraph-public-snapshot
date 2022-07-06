import { SeriesChartContent } from '../../../core'

const getYValue = (datum: MockSeriesDatum): number => datum.value
const getXValue = (datum: MockSeriesDatum): Date => new Date(datum.x)

interface MockSeriesDatum {
    value: number
    x: number
}

export const SERIES_MOCK_CHART: SeriesChartContent<MockSeriesDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, value: 20 },
                { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, value: 40 },
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 110 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 105 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 160 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 184 },
                { x: 1588965700286, value: 200 },
            ],
            name: 'Go 1.11',
            color: 'var(--oc-indigo-7)',
            getYValue,
            getXValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, value: 200 },
                { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, value: 177 },
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 150 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 165 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 100 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 85 },
                { x: 1588965700286, value: 50 },
            ],
            name: 'Go 1.12',
            color: 'var(--oc-orange-7)',
            getYValue,
            getXValue,
        },
    ],
}
