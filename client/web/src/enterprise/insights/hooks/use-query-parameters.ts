import { useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { useDistinctValue } from './use-distinct-value'

type ParametersResult<V extends Readonly<string[]>> = {
    [K in V[number]]: string | undefined
}

export function useQueryParameters<Keys extends Readonly<string[]>>(keys: Keys): ParametersResult<Keys> {
    const { search } = useLocation()
    const parameters = useDistinctValue(keys)

    return useMemo(() => {
        const parsedQuery = new URLSearchParams(search)

        return parameters.reduce((store, key) => {
            store[key as Keys[number]] = parsedQuery.get(key) ?? undefined

            return store
        }, {} as ParametersResult<Keys>)
    }, [search, parameters])
}
