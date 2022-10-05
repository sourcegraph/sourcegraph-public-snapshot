/* eslint-disable react/no-array-index-key */
import React from 'react'

import { MockedResponse } from '@apollo/client/testing'
import { render, waitFor } from '@testing-library/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { SearchHistoryEventLogsQueryResult } from '../../graphql-operations'

import { SEARCH_HISTORY_EVENT_LOGS_QUERY, useRecentSearches } from './recentSearches'

function buildMockTempSettings(items: number): RecentSearch[] {
    return Array.from({ length: items }, (_item, index) => ({
        query: `test${index}`,
        timestamp: '2021-01-01T00:00:00Z',
    }))
}

function buildMockEventLogs(items: number): SearchHistoryEventLogsQueryResult {
    return {
        currentUser: {
            __typename: 'User',
            recentSearchLogs: {
                nodes: Array.from({ length: items }, (_item, index) => ({
                    argument: `{"code_search": { "query_data": { "combined": "test${index}"}}}`,
                    timestamp: '2021-01-01T00:00:00Z',
                })),
            },
        },
    }
}

const Wrapper: React.FunctionComponent<{
    tempSettings: RecentSearch[]
    eventLogs: SearchHistoryEventLogsQueryResult
}> = ({ tempSettings, eventLogs }) => {
    const mockedEventLogs: MockedResponse[] = [
        {
            request: { query: getDocumentNode(SEARCH_HISTORY_EVENT_LOGS_QUERY), variables: { first: 20 } },
            result: { data: eventLogs },
        },
    ]

    return (
        <MockedTestProvider mocks={mockedEventLogs}>
            <MockTemporarySettings settings={{ 'search.input.recentSearches': tempSettings }}>
                <InnerWrapper />
            </MockTemporarySettings>
        </MockedTestProvider>
    )
}

const InnerWrapper: React.FunctionComponent<{}> = () => {
    const { recentSearches, state } = useRecentSearches()
    return (
        <>
            <ul>
                {recentSearches?.map((recentSearch, index) => (
                    <li key={index}>{recentSearch.query}</li>
                ))}
            </ul>
            <div data-testid="state">{state}</div>
        </>
    )
}

describe('recentSearches', () => {
    describe('useRecentSearches().recentSearches', () => {
        test('recent searches is empty array if no data in temp settings or event logs', async () => {
            const { queryAllByRole, getByTestId } = render(
                <Wrapper tempSettings={buildMockTempSettings(0)} eventLogs={buildMockEventLogs(0)} />
            )

            await waitFor(() => expect(getByTestId('state')).toHaveTextContent('success'))

            const items = queryAllByRole('listitem').map(element => element.textContent)
            expect(items).toMatchInlineSnapshot('Array []')
        })

        test('recent searches is populated from event logs if no data in temp settings, with deduplication', async () => {
            const mockedEventLogs = buildMockEventLogs(2)
            const nodes = mockedEventLogs.currentUser?.recentSearchLogs?.nodes ?? []
            const mockedEventLogsWithDuplicates: SearchHistoryEventLogsQueryResult = {
                currentUser: {
                    __typename: 'User',
                    recentSearchLogs: {
                        nodes: [...nodes, ...nodes],
                    },
                },
            }

            const { queryAllByRole, getByTestId } = render(
                <Wrapper tempSettings={buildMockTempSettings(0)} eventLogs={mockedEventLogsWithDuplicates} />
            )

            await waitFor(() => expect(getByTestId('state')).toHaveTextContent('success'))

            const items = queryAllByRole('listitem').map(element => element.textContent)
            expect(items).toMatchInlineSnapshot(
                `
                Array [
                  "test0",
                  "test1",
                ]
            `
            )
        })

        test('recent searches is populated from temp settings', () => {})

        test('adding item to recent searches puts it at the top', () => {})
    })

    describe('useRecentSearches().addRecentSearch', () => {
        test('adding item to recent searches puts it at the top', () => {})

        test('adding an exisitng item to recent searches deduplicates it and puts it at the top', () => {})

        test('adding an item beyond the limit of the list removes the last item', () => {})
    })

    describe('searchHistorySource', () => {
        test('returns null if no recent searches', () => {})

        test('returns recent searches in the correct format', () => {})
    })
})
