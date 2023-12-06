import type { MockedResponse } from '@apollo/client/testing/core'
import type { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import {
    type GetSearchAggregationResult,
    SearchAggregationMode,
    SearchPatternType,
} from '../../../../graphql-operations'

import { AGGREGATION_SEARCH_QUERY } from './hooks'
import { SearchAggregationResult } from './SearchAggregationResult'

const config: Meta = {
    title: 'web/search/results/SearchAggregationResult',
    parameters: {
        chromatic: { disableSnapshots: false },
    },
}

export default config

const SEARCH_AGGREGATION_MOCK: MockedResponse<GetSearchAggregationResult> = {
    request: {
        query: getDocumentNode(AGGREGATION_SEARCH_QUERY),
        variables: {
            query: '',
            patternType: 'literal',
            mode: null,
            limit: 30,
            skipAggregation: false,
        },
    },
    result: {
        data: {
            searchQueryAggregate: {
                __typename: 'SearchQueryAggregate',
                aggregations: {
                    __typename: 'NonExhaustiveSearchAggregationResult',
                    mode: SearchAggregationMode.CAPTURE_GROUP,
                    approximateOtherGroupCount: 776,
                    groups: [
                        {
                            __typename: 'AggregationGroup',
                            label: ': ',
                            count: 232,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:: )string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: ' ',
                            count: 215,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?: )string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: '',
                            count: 68,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:)string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: '?: ',
                            count: 43,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:\\?: )string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: 'ExampleFrom',
                            count: 38,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:ExampleFrom)string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: '       ',
                            count: 26,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:       )string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: ', summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []',
                            count: 15,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:, summaryQuery, err := makeEventLogsQueries\\(s\\.DateRange, s\\.Grouping, \\[\\])string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: '.',
                            count: 11,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:\\.)string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: '  ',
                            count: 10,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:  )string/',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: 'Parameters: map[string][]',
                            count: 8,
                            query: 'context:global repo:^github\\.com/sourcegraph/sourcegraph$ /query(?:Parameters: map\\[string\\]\\[\\])string/',
                        },
                    ],
                },
                modeAvailability: [
                    {
                        __typename: 'AggregationModeAvailability',
                        mode: SearchAggregationMode.REPO,
                        available: true,
                        reasonUnavailable: null,
                    },
                    {
                        __typename: 'AggregationModeAvailability',
                        mode: SearchAggregationMode.PATH,
                        available: true,
                        reasonUnavailable: null,
                    },
                    {
                        __typename: 'AggregationModeAvailability',
                        mode: SearchAggregationMode.AUTHOR,
                        available: false,
                        reasonUnavailable: 'Author aggregation mode is unavailable',
                    },
                    {
                        __typename: 'AggregationModeAvailability',
                        mode: SearchAggregationMode.CAPTURE_GROUP,
                        available: false,
                        reasonUnavailable: 'Capture group aggregation mode is unavailable',
                    },
                ],
            },
        },
    },
}

export const SearchAggregationResultDemo: Story = () => (
    <BrandedStory>
        {() => (
            <MockedTestProvider mocks={[SEARCH_AGGREGATION_MOCK]}>
                <SearchAggregationResult
                    query=""
                    patternType={SearchPatternType.literal}
                    caseSensitive={false}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                    onQuerySubmit={noop}
                />
            </MockedTestProvider>
        )}
    </BrandedStory>
)
