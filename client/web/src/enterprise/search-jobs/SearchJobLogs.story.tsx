import { MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client/out/src'
import { MockedTestProvider } from '@sourcegraph/shared/out/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { GetSearchJobLogsResult, GetSearchJobLogsVariables } from '../../graphql-operations'

import { SEARCH_JOB_LOGS, SearchJobLogs } from './SearchJobLogs'

const defaultStory: Meta = {
    title: 'web/search-jobs',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default defaultStory

const SEARCH_JOB_LOGS_MOCK: MockedResponse<GetSearchJobLogsResult, GetSearchJobLogsVariables> = {
    request: {
        query: getDocumentNode(SEARCH_JOB_LOGS),
        variables: {
            id: '001',
            first: 50,
            after: null,
        },
    },
    result: {
        data: {
            __typename: 'Query',
            searchJob: {
                __typename: 'SearchJob',
                logs: {
                    nodes: [
                        { time: '2023-09-12T20:42:46Z', text: 'test log #1' },
                        { time: '2023-09-12T20:43:46Z', text: 'test log #2' },
                        { time: '2023-09-12T20:44:46Z', text: 'test log #3' },
                        { time: '2023-09-12T20:45:46Z', text: 'test log #4' },
                        { time: '2023-09-12T20:46:46Z', text: 'test log #5' },
                        { time: '2023-09-12T20:47:46Z', text: 'test log #6' },
                        { time: '2023-09-12T20:48:46Z', text: 'test log #7' },
                        { time: '2023-09-12T20:49:46Z', text: 'test log #8' },
                        { time: '2023-09-12T20:50:46Z', text: 'test log #9' },
                        {
                            time: '2023-09-12T20:51:46Z',
                            text: 'Test log #10 and very very loong loooooooong status about the process which happened in that moment',
                        },
                        { time: '2023-09-12T20:52:46Z', text: 'test log #11' },
                        { time: '2023-09-12T20:53:46Z', text: 'test log #12' },
                        { time: '2023-09-12T20:54:46Z', text: 'test log #13' },
                        { time: '2023-09-12T20:55:46Z', text: 'test log #14' },
                        { time: '2023-09-12T20:56:46Z', text: 'test log #15' },
                    ],
                    pageInfo: {
                        hasNextPage: true,
                        endCursor: '002',
                    },
                    totalCount: 16,
                },
            },
        },
    },
}

export const SearchJobLogsContent: Story = () => (
    <MockedTestProvider mocks={[SEARCH_JOB_LOGS_MOCK]}>
        <div
            // Emulate popover content width on the search jobs list page
            style={{ width: '25rem' }}
        >
            <SearchJobLogs jobId="001" onContentChange={() => {}} />
        </div>
    </MockedTestProvider>
)
