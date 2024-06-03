import {
    useQuery,
    useMutation,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { UpdateSubscriptionRequest, Subscription } from '../teamSubscriptions'

import { callCodyProApi } from './callCodyProApi'

// Use query key factories to re-use produced query keys in queries and mutations.
// Motivation taken from here: https://tkdodo.eu/blog/effective-react-query-keys#use-query-key-factories
const queryKeys = {
    all: ['subscription'] as const,
    subscription: () => [...queryKeys.all, 'current-subscription'] as const,
    subscriptionSummary: () => [...queryKeys.all, 'current-subscription-summary'] as const,
}

export const useCurrentSubscription = (): UseQueryResult<Subscription | undefined> =>
    useQuery({
        queryKey: queryKeys.subscription(),
        queryFn: async () => callCodyProApi(Client.getCurrentSubscription()),
    })

export const useUpdateCurrentSubscription = (): UseMutationResult<
    Subscription | undefined,
    Error,
    UpdateSubscriptionRequest
> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => callCodyProApi(Client.updateCurrentSubscription(requestBody)),
        onSuccess: data => {
            // We get updated subscription data in response - no need to refetch subscription.
            // All the `queryKeys.subscription()` subscribers (`useCurrentSubscription` callers) will get the updated value automatically.
            queryClient.setQueryData(queryKeys.subscription(), data)

            // Invalidate `queryKeys.subscriptionSummary()` queries. If the subscription summary is a subset of subscription, we can
            // derive the updated subscription summary from the subscription response eliminating the need in subscription summary query invalidation
            // causing data refetching.
            return queryClient.invalidateQueries({ queryKey: queryKeys.subscriptionSummary() })
        },
    })
}
