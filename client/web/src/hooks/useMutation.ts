import { useMemo, useEffect, useCallback, useState } from 'react'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { isErrorGraphQLResult } from '../../../shared/src/graphql/graphql'
import { createAggregateError } from '../../../shared/src/util/errors'
import { requestGraphQL } from '../backend/graphql'

export interface MutationResult<TData> {
    loading: boolean
    data?: TData
    error?: Error
}

/**
 * Trigger a GraphQL mutation and render based on the returned response
 *
 * @param mutation GraphQL mutation query string
 * @param options Additional configuration
 * @returns Callback to trigger mutation and response data from GraphQL
 */
export function useMutation<TData = unknown, TVariables = unknown>(
    mutation: string,
    options?: {
        /** Throw any errors returned in the response. Use if you want to defer error handling to an ErrorBoundary */
        throwGraphQLErrors: boolean
    }
): [(variables: TVariables) => void, MutationResult<TData>] {
    const subscriptions = useMemo(() => new Subscription(), [])

    const [result, setResult] = useState<MutationResult<TData>>({ loading: false })
    const handleResponse = useCallback(
        (partialResult: Partial<MutationResult<TData>>): void =>
            setResult(previous => ({
                ...previous,
                ...partialResult,
            })),
        []
    )

    const submit = useCallback(
        (variables: TVariables) => {
            handleResponse({ loading: true })

            subscriptions.add(
                requestGraphQL<TData, TVariables>(mutation, variables)
                    .pipe(
                        map(response => {
                            const error = isErrorGraphQLResult(response)
                                ? createAggregateError(response.errors)
                                : undefined

                            if (error && options?.throwGraphQLErrors) {
                                throw error
                            }

                            return handleResponse({ data: response.data, error, loading: false })
                        })
                    )
                    .subscribe()
            )
        },
        [subscriptions, mutation, options, handleResponse]
    )

    return [submit, result]
}
