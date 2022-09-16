import { Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { SeriesSortDirection, SeriesSortMode } from '../../../../graphql-operations'
import { SeriesChartContent, InsightExecutionType, InsightType, SearchBasedInsight } from '../../core'
import { GET_INSIGHT_VIEW_GQL } from '../../core/backend/gql-backend'

import { SmartInsightsViewGrid } from './SmartInsightsViewGrid'

const defaultStory: Meta = {
    title: 'web/insights/SmartInsightsViewGridExample',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            enableDarkMode: true,
        },
    },
}

export default defaultStory

const filters = {
    excludeRepoRegexp: '',
    includeRepoRegexp: '',
    context: '',
    seriesDisplayOptions: {
        limit: '20',
        sortOptions: {
            direction: SeriesSortDirection.DESC,
            mode: SeriesSortMode.RESULT_COUNT,
        },
    },
}

const insightsWithManyLines: SearchBasedInsight[] = [
    {
        id: 'searchInsights.insight.Backend_1',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #1',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_2',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_3',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #3',
        series: [
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
        ],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_4',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #4',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_5',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #5',
        series: [
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
        ],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_6',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #6',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
    {
        id: 'searchInsights.insight.Backend_7',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #7',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { days: 1 },
        filters,
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
]

interface SeriesDatum {
    x: number
    value: number
}

const getXValue = (datum: SeriesDatum): Date => new Date(datum.x)
const getYValue = (datum: SeriesDatum): number => datum.value

const LINE_CHART_WITH_HUGE_NUMBER_OF_LINES: SeriesChartContent<SeriesDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800 },
                { x: 1588965700286, value: 12300 },
            ],
            name: 'React functional components',
            color: 'var(--green)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 17000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19000 },
                { x: 1588965700286, value: 17000 },
            ],
            name: 'Class components',
            color: 'var(--orange)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_003',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 12000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 14000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286, value: 8000 },
            ],
            name: 'useTheme adoption',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_004',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286, value: 8500 },
            ],
            name: 'Class without CSS modules',
            color: 'var(--purple)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_005',
            name: '1.11',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 11000 },
            ],
            color: 'var(--oc-grape-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_006',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 23000 },
                { x: 1588965700286, value: 16000 },
            ],
            name: 'Functional components without CSS modules',
            color: 'var(--oc-red-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_007',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286, value: 9000 },
            ],
            name: '1.12',
            color: 'var(--pink)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_008',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 6000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 6000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 7000 },
                { x: 1588965700286, value: 8000 },
            ],
            name: '1.13',
            color: 'var(--oc-violet-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_009',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 5500 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 7000 },
                { x: 1588965700286, value: 12000 },
            ],
            name: '1.14',
            color: 'var(--indigo)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_010',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 3000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 6000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 6500 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 7000 },
                { x: 1588965700286, value: 10000 },
            ],
            name: '1.15',
            color: 'var(--cyan)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_011',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 6500 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 8500 },
                { x: 1588965700286, value: 15000 },
            ],
            name: '1.16',
            color: 'var(--teal)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_012',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 2000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 18000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19500 },
                { x: 1588965700286, value: 19750 },
            ],
            name: '1.17',
            color: 'var(--oc-lime-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_013',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 2000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 18000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19500 },
                { x: 1588965700286, value: 19750 },
            ],
            name: '1.18',
            color: 'var(--yellow)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_014',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 12000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 6000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286, value: 3000 },
            ],
            name: '1.19',
            color: 'var(--oc-lime-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_015',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 22000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 10000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286, value: 11000 },
            ],
            name: '1.20',
            color: 'var(--oc-pink-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_016',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 14500 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 14500 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 14500 },
                { x: 1588965700286, value: 17000 },
            ],
            name: '1.21',
            color: 'var(--oc-red-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_017',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 12000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 13500 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 15500 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 11500 },
                { x: 1588965700286, value: 12000 },
            ],
            name: '1.22',
            color: 'var(--oc-blue-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_018',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 1000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 2000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 3000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 3500 },
                { x: 1588965700286, value: 4000 },
            ],
            name: '1.23',
            color: 'var(--oc-grape-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_019',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 10000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 7000 },
                { x: 1588965700286, value: 4000 },
            ],
            name: '1.24',
            color: 'var(--oc-green-7)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_020',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 1000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 800 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 7000 },
                { x: 1588965700286, value: 700 },
            ],
            name: '1.25',
            color: 'var(--oc-cyan-7)',
            getXValue,
            getYValue,
        },
    ],
}

const LINE_CHART_WITH_MANY_LINES: SeriesChartContent<SeriesDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800 },
                { x: 1588965700286, value: 12300 },
            ],
            name: 'React functional components',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 26000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19000 },
                { x: 1588965700286, value: 17000 },
            ],
            name: 'Class components',
            color: 'var(--warning)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_003',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 12000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 14000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286, value: 8000 },
            ],
            name: 'useTheme adoption',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_004',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 11000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 8000 },
                { x: 1588965700286, value: 8500 },
            ],
            name: 'Class without CSS modules',
            color: 'var(--purple)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_005',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 63000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 13000 },
                { x: 1588965700286, value: 16000 },
            ],
            name: 'Functional components without CSS modules',
            color: 'var(--green)',
            getXValue,
            getYValue,
        },
    ],
}

const LINE_CHART_TESTS_CASES_EXAMPLE: SeriesChartContent<SeriesDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800 },
                { x: 1588965700286, value: 12300 },
            ],
            name: 'React Test renderer',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_002',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 26000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19000 },
                { x: 1588965700286, value: 17000 },
            ],
            name: 'Enzyme',
            color: 'var(--pink)',
            getXValue,
            getYValue,
        },
        {
            id: 'series_003',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 12000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 14000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 15000 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9000 },
                { x: 1588965700286, value: 8000 },
            ],
            name: 'React Testing Library',
            color: 'var(--red)',
            getXValue,
            getYValue,
        },
    ],
}

const LINE_CHART_TESTS_CASES_SINGLE_EXAMPLE: SeriesChartContent<SeriesDatum> = {
    series: [
        {
            id: 'series_001',
            data: [
                { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000 },
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600 },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800 },
                { x: 1588965700286, value: 12300 },
            ],
            name: 'React Test renderer',
            color: 'var(--blue)',
            getXValue,
            getYValue,
        },
    ],
}

function generateSeries(insight: SearchBasedInsight) {
    const seriesData = getTestCases(insight.series.length)

    return seriesData.series.map(series => ({
        seriesId: series.id,
        label: series.name,
        points: series.data.map(point => ({
            dateTime: new Date(point.x).toUTCString(),
            value: point.value,
            __typename: 'InsightDataPoint',
        })),
        status: {
            backfillQueuedAt: '2021-06-06T15:48:11Z',
            completedJobs: 0,
            pendingJobs: 0,
            failedJobs: 0,
            __typename: 'InsightSeriesStatus',
        },
        __typename: 'InsightsSeries',
    }))
}

function generateMocks(insights: SearchBasedInsight[]) {
    return insights.map(insight => ({
        request: {
            query: GET_INSIGHT_VIEW_GQL,
            variables: {
                id: insight.id,
                filters: { includeRepoRegex: '', excludeRepoRegex: '', searchContexts: [''] },
                seriesDisplayOptions: {
                    limit: 20,
                    sortOptions: { direction: 'DESC', mode: 'RESULT_COUNT' },
                },
            },
        },
        result: {
            data: {
                insightViews: {
                    nodes: [
                        {
                            id: insight.id,
                            appliedSeriesDisplayOptions: {
                                limit: 20,
                                sortOptions: {
                                    mode: 'RESULT_COUNT',
                                    direction: 'DESC',
                                    __typename: 'SeriesSortOptions',
                                },
                                __typename: 'SeriesDisplayOptions',
                            },
                            defaultSeriesDisplayOptions: {
                                limit: null,
                                sortOptions: {
                                    mode: null,
                                    direction: null,
                                    __typename: 'SeriesSortOptions',
                                },
                                __typename: 'SeriesDisplayOptions',
                            },
                            dataSeries: generateSeries(insight),
                            __typename: 'InsightView',
                        },
                    ],
                    __typename: 'InsightViewConnection',
                },
            },
        },
    }))
}

function getTestCases(numberOfSeries: number): SeriesChartContent<SeriesDatum> {
    if (numberOfSeries === 1) {
        return LINE_CHART_TESTS_CASES_SINGLE_EXAMPLE
    }

    if (numberOfSeries >= 15) {
        return LINE_CHART_WITH_HUGE_NUMBER_OF_LINES
    }

    if (numberOfSeries >= 6) {
        return LINE_CHART_WITH_MANY_LINES
    }

    if (numberOfSeries < 6) {
        return LINE_CHART_TESTS_CASES_EXAMPLE
    }

    return LINE_CHART_TESTS_CASES_EXAMPLE
}

function prepInsightSeries(insights: SearchBasedInsight[]): SearchBasedInsight[] {
    return insights.map(insight => {
        const seriesData = getTestCases(insight.series.length)

        const series = seriesData.series.map(data => ({
            id: data.id.toString(),
            query: '',
            stroke: data.color,
            name: data.name,
        }))
        insight.series = series

        return insight
    })
}

export const SmartInsightsViewGridExample = (): JSX.Element => {
    const insights = prepInsightSeries(insightsWithManyLines)
    const mocks = generateMocks(insights)

    return (
        <MockedTestProvider mocks={mocks} addTypename={true}>
            <SmartInsightsViewGrid insights={insights} telemetryService={NOOP_TELEMETRY_SERVICE} />
        </MockedTestProvider>
    )
}
