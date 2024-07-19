import { MockedProvider, type MockedResponse } from '@apollo/client/testing'
import { renderHook } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { type ViewerAffiliatedNamespacesResult, type ViewerAffiliatedNamespacesVariables } from '../graphql-operations'

import { viewerAffiliatedNamespacesMock } from './graphql.mocks'
import { useAffiliatedNamespaces, viewerAffiliatedNamespacesQuery } from './useAffiliatedNamespaces'

describe('useAffiliatedNamespaces', () => {
    test('fetches namespaces', async () => {
        const { result } = renderHook(() => useAffiliatedNamespaces(), {
            wrapper: ({ children }) => (
                <MockedProvider mocks={[viewerAffiliatedNamespacesMock]}>{children}</MockedProvider>
            ),
        })
        await waitForNextApolloResponse()

        expect(result.current.namespaces?.map(ns => ns.id)).toEqual(['user1', 'org1', 'org2'])
        expect(result.current.initialNamespace?.id).toEqual('user1')
        expect(result.current.error).toBeUndefined()
    })

    test('initialNamespaceID', async () => {
        const { result } = renderHook(() => useAffiliatedNamespaces('org2'), {
            wrapper: ({ children }) => (
                <MockedProvider mocks={[viewerAffiliatedNamespacesMock]}>{children}</MockedProvider>
            ),
        })
        await waitForNextApolloResponse()

        expect(result.current.namespaces?.map(ns => ns.id)).toEqual(['user1', 'org1', 'org2'])
        expect(result.current.initialNamespace?.id).toEqual('org2')
        expect(result.current.error).toBeUndefined()
    })

    test('anonymous visitor', async () => {
        const anonymousVisitorAffiliatedNamespacesMock: MockedResponse<
            ViewerAffiliatedNamespacesResult,
            ViewerAffiliatedNamespacesVariables
        > = {
            request: { query: getDocumentNode(viewerAffiliatedNamespacesQuery) },
            result: {
                data: {
                    viewer: {
                        affiliatedNamespaces: {
                            nodes: [],
                        },
                    },
                },
            },
        }

        const { result } = renderHook(() => useAffiliatedNamespaces(), {
            wrapper: ({ children }) => (
                <MockedProvider mocks={[anonymousVisitorAffiliatedNamespacesMock]}>{children}</MockedProvider>
            ),
        })
        await waitForNextApolloResponse()

        expect(result.current.namespaces?.map(ns => ns.id)).toEqual([])
        expect(result.current.initialNamespace).toEqual(undefined)
        expect(result.current.error).toBeUndefined()
    })
})
