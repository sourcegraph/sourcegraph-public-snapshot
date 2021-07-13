import { Dispatch, SetStateAction, useEffect, useRef, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { ConnectionQueryArguments } from '../ConnectionType'
import { QUERY_KEY } from '../constants'
import { getUrlQuery, parseQueryInt } from '../utils'

interface UsePaginationConnectionOptions {
    useURLQuery?: boolean
}

export const useConnectionVariables = <TVariables>(
    variables: TVariables & ConnectionQueryArguments,
    options: UsePaginationConnectionOptions
): [TVariables & ConnectionQueryArguments, Dispatch<SetStateAction<TVariables & ConnectionQueryArguments>>] => {
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

    useEffect(() => {
        if (options.useURLQuery) {
            const searchFragment = getUrlQuery({
                query: connectionVariables.query,
                first: connectionVariables.first,
                defaultFirst,
                location,
                visible: 0,
            })
            if (searchFragment && location.search !== `?${searchFragment}`) {
                history.replace({
                    search: searchFragment,
                    hash: location.hash,
                })
            }
        }
    }, [connectionVariables, defaultFirst, history, location, options.useURLQuery])

    return [connectionVariables, setConnectionVariables]
}
