import React from 'react'

import { MockedResponse } from '@apollo/client/testing'
import { renderHook, waitFor } from '@testing-library/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { MockedFeatureFlagsProvider } from '../../featureFlags/FeatureFlagsProvider'
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
                    argument: `test${index}`,
                    timestamp: '2021-01-01T00:00:00Z',
                })),
            },
        },
    }
}

const createWrapper = (
    featureFlagEnabled: boolean,
    tempSettings: RecentSearch[],
    eventLogs?: SearchHistoryEventLogsQueryResult
) => {
    const mockedEventLogs: MockedResponse[] = eventLogs
        ? [
              {
                  request: { query: getDocumentNode(SEARCH_HISTORY_EVENT_LOGS_QUERY) },
                  result: { data: eventLogs },
              },
          ]
        : []

    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <MockedTestProvider mocks={mockedEventLogs}>
            <MockedFeatureFlagsProvider
                overrides={{ 'search-input-show-history': featureFlagEnabled }}
                refetchInterval={1}
            >
                <MockTemporarySettings settings={{ 'search.input.recentSearches': tempSettings }}>
                    {children}
                </MockTemporarySettings>
            </MockedFeatureFlagsProvider>
        </MockedTestProvider>
    )
    return wrapper
}

describe('recentSearches', () => {
    describe('useRecentSearches().recentSearches', () => {
        test('recent searches is empty array if feature flag is off', async () => {
            const wrapper = createWrapper(false, buildMockTempSettings(5), buildMockEventLogs(5))
            const { result } = renderHook(() => useRecentSearches(), { wrapper })
            await waitFor(() => expect(result.current.recentSearches).toEqual([]))
        })

        test('recent searches is empty array if no data in temp settings or event logs', async () => {
            const wrapper = createWrapper(true, buildMockTempSettings(0), buildMockEventLogs(0))
            const { result } = renderHook(() => useRecentSearches(), { wrapper })
            await waitFor(() => expect(result.current.recentSearches).toEqual([]))
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

            const wrapper = createWrapper(true, buildMockTempSettings(0), mockedEventLogsWithDuplicates)
            const { result } = renderHook(() => useRecentSearches(), { wrapper })
            await waitFor(() =>
                expect(result.current.recentSearches).toEqual([
                    { query: 'test0', timestamp: '2021-01-01T00:00:00Z' },
                    { query: 'test1', timestamp: '2021-01-01T00:00:00Z' },
                ])
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
        test('returns null if feature flag is off', () => {})

        test('returns null if no recent searches', () => {})

        test('returns recent searches in the correct format', () => {})
    })
})
