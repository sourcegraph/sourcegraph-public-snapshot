import {
    useQuery,
    useMutation,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
} from '@tanstack/react-query'

import { type Call, Client, CodyProApiCaller } from '../client'
import type { UpdateSubscriptionRequest, Subscription } from '../teamSubscriptions'

const apiCaller = new CodyProApiCaller()

// Wrapper around `apiCaller.call` which throws an error in case of a non-2xx response.
// Motivation taken from here: https://tanstack.com/query/latest/docs/framework/react/guides/query-functions#usage-with-fetch-and-other-clients-that-do-not-throw-by-default
const callCodyProApi = async <Data>(call: Call<Data>): Promise<{ data?: Data; response: Response }> => {
    const result = await apiCaller.call(call)

    if (!result.response.ok) {
        throw new Error(`Cody Pro API call failed with status ${result.response.status}`)
    }

    // TODO: handle 401 error (e.g., navigate to "/-/sign-out")

    return result
}

// Motivation to use query factories taken from here: https://tkdodo.eu/blog/effective-react-query-keys#use-query-key-factories
const queryKeys = {
    all: ['subscription'] as const,
    subscription: () => [...queryKeys.all, 'current-subscription'] as const,
    subscriptionSummary: () => [...queryKeys.all, 'current-subscription-summary'] as const,
}

export const useCurrentSubscription = (): UseQueryResult<{ data?: Subscription; response: Response }> =>
    useQuery({
        queryKey: queryKeys.subscription(),
        queryFn: async () => callCodyProApi(Client.getCurrentSubscription()),
    })

export const useUpdateCurrentSubscription = (): UseMutationResult<
    { data?: Subscription; response: Response },
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
