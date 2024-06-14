import {
    useQuery,
    useMutation,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
    type UseQueryOptions,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { UpdateSubscriptionRequest, Subscription, SubscriptionSummary } from '../teamSubscriptions'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useCurrentSubscription = (): UseQueryResult<Subscription | undefined> =>
    useQuery({
        queryKey: queryKeys.subscriptions.subscription(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentSubscription())
            return response?.json()
        },
    })

export const useSubscriptionSummary = (
    options: Omit<
        UseQueryOptions<
            SubscriptionSummary | undefined,
            Error,
            SubscriptionSummary | undefined,
            ReturnType<typeof queryKeys.subscriptions.subscriptionSummary>
        >,
        'queryKey' | 'queryFn'
    > = {}
): UseQueryResult<SubscriptionSummary | undefined> =>
    useQuery({
        queryKey: queryKeys.subscriptions.subscriptionSummary(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentSubscriptionSummary())
            return response?.json()
        },
        ...options,
    })

export const useUpdateCurrentSubscription = (): UseMutationResult<
    Subscription | undefined,
    Error,
    UpdateSubscriptionRequest
> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => {
            const response = await callCodyProApi(Client.updateCurrentSubscription(requestBody))
            return response?.json()
        },
        onSuccess: data => {
            // We get updated subscription data in response - no need to refetch subscription.
            // All the `queryKeys.subscription()` subscribers (`useCurrentSubscription` callers) will get the updated value automatically.
            queryClient.setQueryData(queryKeys.subscriptions.subscription(), data)

            // Invalidate `queryKeys.subscriptionSummary()` queries. If the subscription summary is a subset of subscription, we can
            // derive the updated subscription summary from the subscription response eliminating the need in subscription summary query invalidation
            // causing data refetching.
            return queryClient.invalidateQueries({ queryKey: queryKeys.subscriptions.subscriptionSummary() })
        },
    })
}

export const useCreateTeam = (): UseMutationResult<void, Error, CreateTeamRequest> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => {
            await callCodyProApi(Client.createTeam(requestBody))
        },
        onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.all }),
    })
}

export const usePreviewCreateTeam = (): UseMutationResult<PreviewResult | undefined, Error, PreviewCreateTeamRequest> =>
    useMutation({
        mutationFn: async requestBody => {
            const response = await callCodyProApi(Client.previewCreateTeam(requestBody))
            return (await response.json()) as PreviewResult
        },
    })
