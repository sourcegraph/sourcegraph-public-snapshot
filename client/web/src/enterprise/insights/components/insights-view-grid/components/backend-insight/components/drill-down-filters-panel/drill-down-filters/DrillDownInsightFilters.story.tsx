import { useState } from 'react'

import type { MockedResponse } from '@apollo/client/testing/core/mocking/mockLink'
import type { Meta, StoryFn } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../../../../../components/WebStory'
import {
    type GetSearchContextsResult,
    SeriesSortDirection,
    SeriesSortMode,
} from '../../../../../../../../../graphql-operations'
import type { InsightFilters } from '../../../../../../../core'
import { SEARCH_CONTEXT_GQL } from '../search-context/DrillDownSearchContextFilter'

import { DrillDownInsightFilters, FilterSectionVisualMode } from './DrillDownInsightFilters'

const defaultStory: Meta = {
    title: 'web/insights/DrillDownInsightFilters',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

const CONTEXTS_GQL_MOCKS: MockedResponse<GetSearchContextsResult> = {
    request: { query: SEARCH_CONTEXT_GQL, variables: { query: '' } },
    error: undefined,
    result: {
        data: {
            __typename: 'Query',
            searchContexts: {
                nodes: [
                    {
                        __typename: 'SearchContext',
                        id: '001',
                        spec: 'global',
                        query: 'repo:github.com/sourcegraph/sourcegraph',
                        description: 'Hello this is mee, your friend context',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '002',
                        spec: 'sourcegraph',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '003',
                        spec: '@sourcegraph/code-insights',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '004',
                        spec: '@sourcegraph/code-insights',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '005',
                        spec: '@sourcegraph/code-insights',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '006',
                        spec: '@sourcegraph/code-insights',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                ],
                pageInfo: {
                    hasNextPage: false,
                },
            },
        },
    },
}

const ORIGINAL_FILTERS: InsightFilters = {
    includeRepoRegexp: '',
    excludeRepoRegexp: '',
    context: '',
    seriesDisplayOptions: {
        limit: 20,
        numSamples: null,
        sortOptions: {
            direction: SeriesSortDirection.DESC,
            mode: SeriesSortMode.RESULT_COUNT,
        },
    },
}

const FILTERS: InsightFilters = {
    includeRepoRegexp: 'hello world loooong loooooooooooooong repo filter regular expressssssion',
    excludeRepoRegexp: 'hello world loooong loooooooooooooong repo filter regular expressssssion',
    context: '',
    seriesDisplayOptions: {
        limit: 20,
        numSamples: null,
        sortOptions: {
            direction: SeriesSortDirection.DESC,
            mode: SeriesSortMode.RESULT_COUNT,
        },
    },
}

export const DrillDownFiltersShowcase: StoryFn = () => (
    <MockedTestProvider mocks={[CONTEXTS_GQL_MOCKS]}>
        <DrillDownInsightFilters
            initialValues={FILTERS}
            originalValues={ORIGINAL_FILTERS}
            isNumSamplesFilterAvailable={true}
            visualMode={FilterSectionVisualMode.CollapseSections}
            onFiltersChange={console.log}
            onFilterSave={console.log}
            onCreateInsightRequest={console.log}
        />
    </MockedTestProvider>
)

export const DrillDownFiltersHorizontalMode: StoryFn = () => {
    const [mode, setMode] = useState<FilterSectionVisualMode>(FilterSectionVisualMode.HorizontalSections)

    return (
        <MockedTestProvider mocks={[CONTEXTS_GQL_MOCKS]}>
            <DrillDownInsightFilters
                initialValues={FILTERS}
                originalValues={ORIGINAL_FILTERS}
                isNumSamplesFilterAvailable={true}
                visualMode={mode}
                onVisualModeChange={setMode}
                onFiltersChange={console.log}
                onFilterSave={console.log}
                onCreateInsightRequest={console.log}
            />
        </MockedTestProvider>
    )
}
