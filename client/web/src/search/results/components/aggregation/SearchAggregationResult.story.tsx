import { MockedResponse } from '@apollo/client/testing/core'
import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { getDocumentNode } from '@sourcegraph/http-client'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { GetSearchAggregationResult, SearchAggregationMode } from '../../../../graphql-operations'

import { AGGREGATION_SEARCH_QUERY } from './hooks'
import { SearchAggregationResult } from './SearchAggregationResult'

const config: Meta = {
    title: 'web/search/results/SearchAggregationResult',
}

export default config

const SEARCH_AGGREGATION_MOCK: MockedResponse<GetSearchAggregationResult> = {
    request: {
        query: getDocumentNode(AGGREGATION_SEARCH_QUERY),
        variables: {
            query: '',
            patternType: 'literal',
            mode: 'REPO',
            limit: 30,
        },
    },
    result: {
        data: {
            searchQueryAggregate: {
                __typename: 'SearchQueryAggregate',
                aggregations: {
                    __typename: 'ExhaustiveSearchAggregationResult',
                    mode: SearchAggregationMode.REPO,
                    otherGroupCount: 100,
                    groups: [
                        {
                            __typename: 'AggregationGroup',
                            label: 'sourcegraph/sourcegraph',
                            count: 100,
                            query: 'context:global insights repo:sourcegraph/sourcegraph',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: 'sourcegraph/about',
                            count: 80,
                            query: 'context:global insights repo:sourecegraph/about',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: 'sourcegraph/search-insight',
                            count: 60,
                            query: 'context:global insights repo:sourecegraph/search-insight',
                        },
                        {
                            __typename: 'AggregationGroup',
                            label: 'sourcegraph/lang-stats',
                            count: 40,
                            query: 'context:global insights repo:sourecegraph/lang-stats',
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
                    onQuerySubmit={noop}
                />
            </MockedTestProvider>
        )}
    </BrandedStory>
)
