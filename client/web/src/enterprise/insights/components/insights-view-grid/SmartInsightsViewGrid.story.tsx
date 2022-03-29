import React from 'react'

import { Meta } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import {
    LINE_CHART_TESTS_CASES_EXAMPLE,
    LINE_CHART_WITH_HUGE_NUMBER_OF_LINES,
    LINE_CHART_WITH_MANY_LINES,
} from '../../../../views/mocks/charts-content'
import { CodeInsightsBackendStoryMock } from '../../CodeInsightsBackendStoryMock'
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

const codeInsightsApiWithManyLines = {
    getBackendInsightData: (insight: BackendInsight) => {
        if (isCaptureGroupInsight(insight)) {
            throw new Error('This demo does not support capture group insight')
        }

        return of({
            id: insight.id,
            view: {
                title: 'Backend Insight Mock',
                subtitle: 'Backend insight description text',
                content: [
                    insight.series.length >= 6
                        ? insight.series.length >= 15
                            ? LINE_CHART_WITH_HUGE_NUMBER_OF_LINES
                            : LINE_CHART_WITH_MANY_LINES
                        : LINE_CHART_TESTS_CASES_EXAMPLE,
                ],
                isFetchingHistoricalData: false,
            },
        })
    },
}

export const SmartInsightsViewGridExample = (): JSX.Element => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsApiWithManyLines}>
        <SmartInsightsViewGrid insights={insightsWithManyLines} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendStoryMock>
)
