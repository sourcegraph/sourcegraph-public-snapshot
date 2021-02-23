import { useMemo, useCallback, useState } from 'react'
import { Subscription } from 'rxjs'
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
    const submit = useCallback(
        (variables: TVariables) => {
            setResult({ loading: true, data: undefined, error: undefined })

            subscriptions.add(
                requestGraphQL<TData, TVariables>(mutation, variables).subscribe(response => {
                    const error = isErrorGraphQLResult(response) ? createAggregateError(response.errors) : undefined
                    setResult({ loading: false, data: response.data, error })
                })
            )
        },
        [subscriptions, mutation]
    )

    if (result.error && options?.throwGraphQLErrors) {
        throw result.error
    }

    return [submit, result]
}
