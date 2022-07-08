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

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
}

export const COMPUTE_MOCK_CHART: LanguageUsageDatum[] = [
    {
        name: 'JavaScript',
        value: 422,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'CSS',
        value: 273,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'HTML',
        value: 20,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'Markdown',
        value: 135,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'JavaScript',
        value: 300,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'CSS',
        value: 150,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'HTML',
        value: 390,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'Markdown',
        value: 300,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]
