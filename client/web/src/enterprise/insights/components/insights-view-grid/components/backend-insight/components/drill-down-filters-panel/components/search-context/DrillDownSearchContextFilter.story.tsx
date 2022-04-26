import React from 'react'

import { MockedResponse } from '@apollo/client/testing/core/mocking/mockLink'
import { Meta } from '@storybook/react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../../../../../../components/WebStory'
import { GetSearchContextsResult } from '../../../../../../../../../../graphql-operations'

import { DrillDownSearchContextFilter, SEARCH_CONTEXT_GQL } from './DrillDownSearchContextFilter'

const defaultStory: Meta = {
    title: 'web/insights/DrillDownSearchContextFilter',
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

export const DrillDownSearchContextFilterExample = () => (
    <MockedTestProvider mocks={[CONTEXTS_GQL_MOCKS]}>
        <DrillDownSearchContextFilter />
    </MockedTestProvider>
)
