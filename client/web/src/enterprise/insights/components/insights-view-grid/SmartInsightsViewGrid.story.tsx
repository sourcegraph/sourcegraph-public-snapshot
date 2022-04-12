import React from 'react'

import { Meta } from '@storybook/react'
import { Observable, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../CodeInsightsBackendStoryMock'
import { BackendInsightData, SeriesChartContent } from '../../core'
import { BackendInsight, Insight, InsightExecutionType, InsightType, isCaptureGroupInsight } from '../../core/types'

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
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_2',
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: 'Backend insight #3',
        series: [],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_3',
        executionType: InsightExecutionType.Backend,
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
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_4',
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_5',
        executionType: InsightExecutionType.Backend,
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
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_6',
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
    {
        id: 'searchInsights.insight.Backend_7',
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: 'Backend insight #2',
        series: [{ id: '', query: '', stroke: '', name: '' }],
        step: { weeks: 2 },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
        dashboardReferenceCount: 0,
        isFrozen: false,
    },
]

interface HugeLinesDatum {
    x: number
    a: number
    b: number
    c: number
    d: number
    e: number | null
    f: number
    g: number
    h: number
    i: number
    j: number
    k: number
    l: number
    m: number
    n: number
    o: number
    p: number
    q: number
    r: number
    s: number
    t: number
}

const LINE_CHART_WITH_HUGE_NUMBER_OF_LINES: SeriesChartContent<HugeLinesDatum> = {
    data: [
        {
            x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
            a: 4000,
            b: 15000,
            c: 12000,
            d: 11000,
            e: 11000,
            f: 13000,
            g: 5000,
            h: 5000,
            i: 5000,
            j: 7000,
            k: 10000,
            l: 8000,
            m: 3900,
            n: 3000,
            o: 4000,
            p: 5000,
            q: 4500,
            r: 5000,
            s: 5500,
            t: 6000,
        },
        {
            x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
            a: 4000,
            b: 17000,
            c: 14000,
            d: 11000,
            e: 11000,
            f: 5000,
            g: 5000,
            h: 6000,
            i: 5500,
            j: 7200,
            k: 8000,
            l: 7800,
            m: 4000,
            n: 3000,
            o: 4500,
            p: 5500,
            q: 5500,
            r: 6000,
            s: 7500,
            t: 5000,
        },
        {
            x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
            a: 5600,
            b: 20000,
            c: 15000,
            d: 13000,
            e: null,
            f: 23000,
            g: 8000,
            h: 7000,
            i: 4500,
            j: 11000,
            k: 10000,
            l: 9000,
            m: 5000,
            n: 3000,
            o: 4000,
            p: 5000,
            q: 4500,
            r: 5000,
            s: 5500,
            t: 6000,
        },
        {
            x: 1588965700286 - 1 * 24 * 60 * 60 * 1000,
            a: 9800,
            b: 19000,
            c: 9000,
            d: 8000,
            e: null,
            f: 13000,
            g: 5000,
            h: 6000,
            i: 5500,
            j: 7200,
            k: 8000,
            l: 7800,
            m: 4000,
            n: 4000,
            o: 5000,
            p: 4000,
            q: 7500,
            r: 8000,
            s: 8500,
            t: 4000,
        },
        {
            x: 1588965700286,
            a: 12300,
            b: 17000,
            c: 8000,
            d: 8500,
            e: null,
            f: 16000,
            g: 9000,
            h: 8000,
            i: 5500,
            j: 12000,
            k: 11000,
            l: 10000,
            m: 6000,
            n: 6000,
            o: 7000,
            p: 8000,
            q: 6500,
            r: 9000,
            s: 10500,
            t: 16000,
        },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'React functional components',
            color: 'var(--green)',
        },
        {
            dataKey: 'b',
            name: 'Class components',
            color: 'var(--orange)',
        },
        { dataKey: 'c', name: 'useTheme adoption', color: 'var(--blue)' },
        { dataKey: 'd', name: 'Class without CSS modules', color: 'var(--purple)' },
        { dataKey: 'e', name: '1.11', color: 'var(--oc-grape-7)' },
        { dataKey: 'f', name: 'Functional components without CSS modules', color: 'var(--oc-red-7)' },
        { dataKey: 'g', name: '1.12', color: 'var(--pink)' },
        { dataKey: 'h', name: '1.13', color: 'var(--oc-violet-7)' },
        { dataKey: 'i', name: '1.14', color: 'var(--indigo)' },
        { dataKey: 'm', name: '1.15', color: 'var(--cyan)' },
        { dataKey: 'j', name: '1.16', color: 'var(--teal)' },
        { dataKey: 'k', name: '1.17', color: 'var(--oc-lime-7)' },
        { dataKey: 'l', name: '1.18', color: 'var(--yellow)' },
        { dataKey: 'n', name: '1.19', color: 'var(--oc-lime-7)' },
        { dataKey: 'o', name: '1.20', color: 'var(--oc-pink-7)' },
        { dataKey: 'p', name: '1.21', color: 'var(--oc-red-7)' },
        { dataKey: 'q', name: '1.22', color: 'var(--oc-blue-7)' },
        { dataKey: 'r', name: '1.23', color: 'var(--oc-grape-7)' },
        { dataKey: 's', name: '1.24', color: 'var(--oc-green-7)' },
        { dataKey: 't', name: '1.25', color: 'var(--oc-cyan-7)' },
    ],
    getXValue: datum => new Date(datum.x),
}

interface ManyLinesDatum {
    x: number
    a: number
    b: number
    c: number
    d: number
    f: number
}

const LINE_CHART_WITH_MANY_LINES: SeriesChartContent<ManyLinesDatum> = {
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 4000, b: 15000, c: 12000, d: 11000, f: 13000 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 4000, b: 26000, c: 14000, d: 11000, f: 5000 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 5600, b: 20000, c: 15000, d: 13000, f: 63000 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 9800, b: 19000, c: 9000, d: 8000, f: 13000 },
        { x: 1588965700286, a: 12300, b: 17000, c: 8000, d: 8500, f: 16000 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'React functional components',
            color: 'var(--warning)',
        },
        {
            dataKey: 'b',
            name: 'Class components',
            color: 'var(--warning)',
        },
        { dataKey: 'c', name: 'useTheme adoption', color: 'var(--blue)' },
        { dataKey: 'd', name: 'Class without CSS modules', color: 'var(--purple)' },
        { dataKey: 'f', name: 'Functional components without CSS modules', color: 'var(--green)' },
    ],
    getXValue: datum => new Date(datum.x),
}

interface TestCasesDatum {
    x: number
    a: number
    b: number
    c: number
    d: number
    f: number
}

const LINE_CHART_TESTS_CASES_EXAMPLE: SeriesChartContent<TestCasesDatum> = {
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 4000, b: 15000, c: 12000, d: 11000, f: 13000 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 4000, b: 26000, c: 14000, d: 11000, f: 5000 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 5600, b: 20000, c: 15000, d: 13000, f: 63000 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 9800, b: 19000, c: 9000, d: 8000, f: 13000 },
        { x: 1588965700286, a: 12300, b: 17000, c: 8000, d: 8500, f: 16000 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'React Test renderer',
            color: 'var(--blue)',
        },
        {
            dataKey: 'b',
            name: 'Enzyme',
            color: 'var(--pink)',
        },
        {
            dataKey: 'c',
            name: 'React Testing Library',
            color: 'var(--red)',
        },
    ],
    getXValue: datum => new Date(datum.x),
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
