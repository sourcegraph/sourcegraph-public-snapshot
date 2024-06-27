import {
    useQuery,
    useMutation,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type {
    CreateTeamRequest,
    PreviewCreateTeamRequest,
    PreviewResult,
    PreviewUpdateSubscriptionRequest,
    Subscription,
    SubscriptionSummary,
    UpdateSubscriptionRequest,
    GetSubscriptionInvoicesResponse,
    ListTeamMembersResponse,
    ListTeamInvitesResponse,
} from '../types'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useCurrentSubscription = (): UseQueryResult<Subscription | undefined> =>
    useQuery({
        queryKey: queryKeys.subscriptions.subscription(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentSubscription())
            return response.json()
        },
    })

export const useSubscriptionSummary = (): UseQueryResult<SubscriptionSummary | undefined> =>
    useQuery({
        queryKey: queryKeys.subscriptions.subscriptionSummary(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentSubscriptionSummary())
            return response.json()
        },
    })

export const useSubscriptionInvoices = (): UseQueryResult<GetSubscriptionInvoicesResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.subscriptions.subscriptionInvoices(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentSubscriptionInvoices())
            return response.json()
        },
    })

export const useTeamMembers = (): UseQueryResult<ListTeamMembersResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.teams.teamMembers(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentTeamMembers())
            return response.ok ? response.json() : undefined
        },
    })

export const useTeamInvites = (): UseQueryResult<ListTeamInvitesResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.invites.teamInvites(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getTeamInvites())
            return response.ok ? response.json() : undefined
        },
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
            return response.json()
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
        onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.subscriptions.all }),
    })
}

export const usePreviewCreateTeam = (): UseMutationResult<PreviewResult | undefined, Error, PreviewCreateTeamRequest> =>
    useMutation({
        mutationFn: async requestBody => {
            const response = await callCodyProApi(Client.previewCreateTeam(requestBody))
            return response.json()
        },
    })

export const usePreviewUpdateCurrentSubscription = (): UseMutationResult<
    PreviewResult | undefined,
    Error,
    PreviewUpdateSubscriptionRequest
> =>
    useMutation({
        mutationFn: async requestBody => {
            const response = await callCodyProApi(Client.previewUpdateCurrentSubscription(requestBody))
            return (await response.json()) as PreviewResult
        },
    })
