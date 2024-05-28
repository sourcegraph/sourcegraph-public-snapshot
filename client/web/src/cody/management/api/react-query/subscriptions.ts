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

class CodyProApiAuthError extends Error {
    constructor(m: string) {
        super(m)

        Object.setPrototypeOf(this, CodyProApiAuthError.prototype)
    }
}

const isCodyProApiAuthError = (err?: Error): err is CodyProApiAuthError => err instanceof CodyProApiAuthError

/**
 * Returns an error message based on the type of error encountered.
 *
 * If the error is a specific Cody Pro API error (e.g., `CodyProApiAuthError` as result of expired SAMS credentials),
 * it returns the error message associated with that error. Otherwise, it returns a custom caller-provided message.
 *
 * @param err - The error object encountered, if any.
 * @param message - The custom message to return if the error is not a specific Cody Pro API authentication error.
 * @returns The error message to display.
 */
export const getCodyProApiErrorMessage = (err: Error | null, message: string): string => {
    if (!err) {
        return ''
    }

    if (isCodyProApiAuthError(err)) {
        return err.message
    }

    return message
}

// Wrapper around `apiCaller.call` which throws an error in case of a non-2xx response.
// Motivation taken from here: https://tanstack.com/query/latest/docs/framework/react/guides/query-functions#usage-with-fetch-and-other-clients-that-do-not-throw-by-default
const callCodyProApi = async <Data>(call: Call<Data>): Promise<{ data?: Data; response: Response }> => {
    const result = await apiCaller.call(call)

    if (!result.response.ok) {
        if (result.response.status === 401) {
            throw new CodyProApiAuthError('Unauthorized. Please log out and log back in.')
        }
        throw new Error(`Cody Pro API call failed with status ${result.response.status}`)
    }

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
