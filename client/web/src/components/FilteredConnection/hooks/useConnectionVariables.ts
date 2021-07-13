import { useRef, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { useDebounce } from '@sourcegraph/wildcard'

import { ConnectionQueryArguments } from '../ConnectionType'
import { QUERY_KEY } from '../constants'
import { getUrlQuery, parseQueryInt } from '../utils'

interface UsePaginationConnectionOptions {
    useURLQuery?: boolean
}

export const useConnectionVariables = <TVariables>(
    variables: TVariables & ConnectionQueryArguments,
    options: UsePaginationConnectionOptions
): [TVariables & ConnectionQueryArguments, (variables: TVariables & ConnectionQueryArguments) => void] => {
    const history = useHistory()
    const location = useLocation()
    const searchParameters = new URLSearchParams(location.search)
    const { current: defaultFirst } = useRef<number>(variables.first || 15)

    const [connectionVariables, setConnectionVariables] = useState({
        ...variables,
        first: (options.useURLQuery && parseQueryInt(searchParameters, 'first')) || variables.first,
        after: (options.useURLQuery && searchParameters.get('after')) || variables.after,
        query: (options.useURLQuery && searchParameters.get(QUERY_KEY)) || variables.query,
    })

    const searchFragment = useDebounce(
        getUrlQuery({
            query: connectionVariables.query,
            first: connectionVariables.first,
            defaultFirst,
            location,
            visible: 0,
        }),
        200
    )

    if (options.useURLQuery && searchFragment && location.search !== `?${searchFragment}`) {
        history.replace({
            search: searchFragment,
            hash: location.hash,
        })
    }

    return [connectionVariables, setConnectionVariables]
}
