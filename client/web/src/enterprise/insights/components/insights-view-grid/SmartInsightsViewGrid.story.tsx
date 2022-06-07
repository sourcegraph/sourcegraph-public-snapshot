import { Meta } from '@storybook/react'
import { Observable, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../CodeInsightsBackendStoryMock'
import {
    BackendInsightData,
    SeriesChartContent,
    BackendInsight,
    Insight,
    InsightExecutionType,
    InsightType,
    isCaptureGroupInsight,
} from '../../core'

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

const insightsWithManyLines: Insight[] = [
    {
        id: 'searchInsights.insight.Backend_1',
        executionType: InsightExecutionType.Backend,
        repositories: [],
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #3',
        series: [],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #1',
        series: [
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
            { id: '', query: '', stroke: '', name: '' },
        ],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #2',
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
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
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
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '', context: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
        seriesDisplayOptions: {},
        dashboards: [],
    },
]

interface SeriesDatum {
    x: number
    value: number | null
}

const getXValue = (datum: SeriesDatum): Date => new Date(datum.x)
const getYValue = (datum: SeriesDatum): number | null => datum.value

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
                { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: null },
                { x: 1588965700286, value: null },
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

const codeInsightsApiWithManyLines = {
    getBackendInsightData: (insight: BackendInsight): Observable<BackendInsightData> => {
        if (isCaptureGroupInsight(insight)) {
            throw new Error('This demo does not support capture group insight')
        }

        return of({
            content:
                insight.series.length >= 6
                    ? insight.series.length >= 15
                        ? LINE_CHART_WITH_HUGE_NUMBER_OF_LINES
                        : LINE_CHART_WITH_MANY_LINES
                    : LINE_CHART_TESTS_CASES_EXAMPLE,
            isFetchingHistoricalData: false,
        })
    },
}

export const SmartInsightsViewGridExample = (): JSX.Element => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsApiWithManyLines}>
        <SmartInsightsViewGrid insights={insightsWithManyLines} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendStoryMock>
)
