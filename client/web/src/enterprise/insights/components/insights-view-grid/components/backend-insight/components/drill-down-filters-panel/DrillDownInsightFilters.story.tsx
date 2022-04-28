import { MockedResponse } from '@apollo/client/testing/core/mocking/mockLink'
import { Meta, Story } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../../../../components/WebStory'
import { GetSearchContextsResult } from '../../../../../../../../graphql-operations'
import { InsightFilters } from '../../../../../../core'

import { DrillDownInsightFilters } from './DrillDownInsightFilters'
import { SEARCH_CONTEXT_GQL } from './search-context/DrillDownSearchContextFilter'

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
                        name: 'global',
                        query: 'repo:github.com/sourcegraph/sourcegraph',
                        description: 'Hello this is mee, your friend context',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '002',
                        name: 'sourcegraph',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '003',
                        name: '@sourcegraph/code-insights',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '004',
                        name: 'Test context 2',
                        query: 'repo:github.com/sourcegraph/sourcegraph2',
                        description: 'Hello this is mee, your friend context 2',
                    },
                    {
                        __typename: 'SearchContext',
                        id: '005',
                        name: 'Test context 2',
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
}

const FILTERS: InsightFilters = {
    includeRepoRegexp: 'hello world loooong loooooooooooooong repo filter regular expressssssion',
    excludeRepoRegexp: 'hello world loooong loooooooooooooong repo filter regular expressssssion',
    context: '',
}

export const DrillDownFiltersShowcase: Story = () => (
    <MockedTestProvider mocks={[CONTEXTS_GQL_MOCKS]}>
        <DrillDownInsightFilters
            initialValues={FILTERS}
            originalValues={ORIGINAL_FILTERS}
            onFiltersChange={console.log}
            onFilterSave={console.log}
            onCreateInsightRequest={console.log}
        />
    </MockedTestProvider>
)
