import { useMemo, useEffect, useCallback, useState } from 'react'
import { Subscription, throwError } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { isErrorGraphQLResult } from '../../../shared/src/graphql/graphql'
import { asError, createAggregateError } from '../../../shared/src/util/errors'
import { requestGraphQL } from '../backend/graphql'

interface MutationResult<TData> {
    loading: boolean
    data?: TData
    error?: Error
}

export function useMutation<TData = unknown, TVariables = unknown>(
    mutation: string
): [(options: TVariables) => void, MutationResult<TData>] {
    const subscriptions = useMemo(() => new Subscription(), [])
    useEffect(() => () => subscriptions.unsubscribe(), [subscriptions])

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
        (options: TVariables) => {
            handleResponse({ loading: true })

            subscriptions.add(
                requestGraphQL<TData, TVariables>(mutation, options)
                    .pipe(
                        map(response => {
                            const error = isErrorGraphQLResult(response)
                                ? createAggregateError(response.errors)
                                : undefined
                            return handleResponse({ data: response.data, error, loading: false })
                        }),
                        catchError(error => throwError(asError(error)))
                    )
                    .subscribe()
            )
        },
        [subscriptions, mutation, handleResponse]
    )

    return [submit, result]
}
