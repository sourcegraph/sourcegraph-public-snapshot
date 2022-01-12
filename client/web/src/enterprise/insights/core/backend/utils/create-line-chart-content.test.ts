import fetch from 'jest-fetch-mock'

import { SearchBasedBackendFilters, SearchBasedInsightSeries } from '../../types/insight/search-insight'

import {
    createLineChartContent,
    createLineChartContentFromIndexedSeries,
    InsightData,
    InsightDataSeriesData,
} from './create-line-chart-content'

const FILTERS: SearchBasedBackendFilters = {
    includeRepoRegexp: '',
    excludeRepoRegexp: '',
}

const MOCK_SERIES_DATA: InsightData['series'] = [
    {
        label: '#1 line',
        points: [
            {
                dateTime: '2021-11-01T00:47:54Z',
                value: 100,
            },
            {
                dateTime: '2021-10-01T00:33:48Z',
                value: 101,
            },
            {
                dateTime: '2021-09-01T00:33:48Z',
                value: 102,
            },
        ],
    },
    {
        label: '#2 line',
        points: [
            {
                dateTime: '2021-11-01T00:47:54Z',
                value: 200,
            },
            {
                dateTime: '2021-10-01T00:33:48Z',
                value: 201,
            },
            {
                dateTime: '2021-09-02T00:33:48Z',
                value: 202,
            },
        ],
    },
]

const MOCK_SERIES_DEFINITION: SearchBasedInsightSeries[] = [
    {
        id: '001',
        query: 'series 1 query',
        name: '#1 line',
        stroke: 'blue',
    },
    {
        id: '002',
        query: 'series 2 query',
        name: '#2 line',
        stroke: 'orange',
    },
]

describe('createLineChartContent', () => {
    it('should generate empty line chart data with no series data', () => {
        expect(createLineChartContent({ series: [] }, [])).toStrictEqual({
            chart: 'line',
            data: [],
            series: [],
            xAxis: {
                dataKey: 'dateTime',
                scale: 'time',
                type: 'number',
            },
        })
    })

    it('should generate char line content with series data within', () => {
        expect(createLineChartContent({ series: MOCK_SERIES_DATA }, MOCK_SERIES_DEFINITION)).toStrictEqual({
            chart: 'line',
            data: [
                {
                    dateTime: 1635727674000,
                    series0: 100,
                    series1: 200,
                },
                {
                    dateTime: 1633048428000,
                    series0: 101,
                    series1: 201,
                },
                {
                    dateTime: 1630456428000,
                    series0: 102,
                    series1: null,
                },
                {
                    dateTime: 1630542828000,
                    series0: null,
                    series1: 202,
                },
            ],
            series: [
                {
                    dataKey: 'series0',
                    name: '#1 line',
                    stroke: 'blue',
                },
                {
                    dataKey: 'series1',
                    name: '#2 line',
                    stroke: 'orange',
                },
            ],
            xAxis: {
                dataKey: 'dateTime',
                scale: 'time',
                type: 'number',
            },
        })
    })
})

const MOCK_INDEXED_SERIES_DATA: InsightDataSeriesData[] = [
    {
        seriesId: '001',
        label: '#1 line',
        points: [
            {
                dateTime: '2021-11-01T00:47:54Z',
                value: 100,
            },
            {
                dateTime: '2021-10-01T00:33:48Z',
                value: 101,
            },
            {
                dateTime: '2021-09-01T00:33:48Z',
                value: 102,
            },
        ],
    },
    {
        seriesId: '002',
        label: '#2 line',
        points: [
            {
                dateTime: '2021-11-01T00:47:54Z',
                value: 200,
            },
            {
                dateTime: '2021-10-01T00:33:48Z',
                value: 201,
            },
            {
                dateTime: '2021-09-02T00:33:48Z',
                value: 202,
            },
        ],
    },
]

describe('createLineChartContentFromIndexedSeries', () => {
    const { location } = window

    beforeAll(() => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        delete window.location
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        window.location = new URL('https://sourcegraph.test')
    })

    beforeEach(() => {
        fetch.enableMocks()
        fetch.mockClear()
    })

    afterAll(() => {
        fetch.disableMocks()

        window.location = location
    })

    it('should generate empty line chart data with no series data', () => {
        expect(createLineChartContentFromIndexedSeries([], [], FILTERS)).toStrictEqual({
            chart: 'line',
            data: [],
            series: [],
            xAxis: {
                dataKey: 'dateTime',
                scale: 'time',
                type: 'number',
            },
        })
    })

    it('should generate char line content with series data within', () => {
        expect(
            createLineChartContentFromIndexedSeries(MOCK_INDEXED_SERIES_DATA, MOCK_SERIES_DEFINITION, FILTERS)
        ).toStrictEqual({
            chart: 'line',
            data: [
                {
                    dateTime: 1635727674000,
                    '001': 100,
                    '002': 200,
                },
                {
                    dateTime: 1633048428000,
                    '001': 101,
                    '002': 201,
                },
                {
                    dateTime: 1630456428000,
                    '001': 102,
                    '002': null,
                },
                {
                    dateTime: 1630542828000,
                    '001': null,
                    '002': 202,
                },
            ],
            series: [
                {
                    dataKey: '001',
                    name: '#1 line',
                    stroke: 'blue',
                    linkURLs: {
                        '1630456428000':
                            'https://sourcegraph.test/search?q=++type%3Adiff++before%3A2021-09-01T00%3A33%3A48Z+series+1+query',
                        '1633048428000':
                            'https://sourcegraph.test/search?q=++type%3Adiff+after%3A2021-09-01T00%3A33%3A48Z+before%3A2021-10-01T00%3A33%3A48Z+series+1+query',
                        '1635727674000':
                            'https://sourcegraph.test/search?q=++type%3Adiff+after%3A2021-10-01T00%3A33%3A48Z+before%3A2021-11-01T00%3A47%3A54Z+series+1+query',
                    },
                },
                {
                    dataKey: '002',
                    name: '#2 line',
                    stroke: 'orange',
                    linkURLs: {
                        '1630542828000':
                            'https://sourcegraph.test/search?q=++type%3Adiff++before%3A2021-09-02T00%3A33%3A48Z+series+2+query',
                        '1633048428000':
                            'https://sourcegraph.test/search?q=++type%3Adiff+after%3A2021-09-02T00%3A33%3A48Z+before%3A2021-10-01T00%3A33%3A48Z+series+2+query',
                        '1635727674000':
                            'https://sourcegraph.test/search?q=++type%3Adiff+after%3A2021-10-01T00%3A33%3A48Z+before%3A2021-11-01T00%3A47%3A54Z+series+2+query',
                    },
                },
            ],
            xAxis: {
                dataKey: 'dateTime',
                scale: 'time',
                type: 'number',
            },
        })
    })
})
