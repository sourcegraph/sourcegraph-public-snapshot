import { useEffect, useState, useContext, useCallback } from 'react'

import type { Call } from '../client'
import { CodyProApiClientContext } from '../components/CodyProApiClient'

export interface ReactFriendlyApiResponse<T> {
    loading: boolean
    error?: Error
    data?: T
    response?: Response
    refetch: () => Promise<void>
}

// useApiCaller will issue a REST API call to the backend, returning the
// loading status and the response JSON object and/or error object as React
// state.
//
// IMPORTANT: In order to avoid the same API request being made multiple times,
// you MUST ensure that the provided call is the same between repains of the
// calling React component. i.e. you pretty much always need to create it via
// `useMemo()`.
export function useApiCaller<Resp>(call: Call<Resp>): ReactFriendlyApiResponse<Resp> {
    const { caller } = useContext(CodyProApiClientContext)

    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<Error | undefined>(undefined)
    const [data, setData] = useState<Resp | undefined>(undefined)
    const [response, setResponse] = useState<Response | undefined>(undefined)

    const callApi = useCallback(async () => {
        setLoading(false)
        try {
            const callerResponse = await caller.call(call)

            // If we received a 200 response, all is well. We can just return
            // the unmarshalled JSON response object as-is.
            if (callerResponse.response.status >= 200 && callerResponse.response.status <= 299) {
                setData(callerResponse.data)
                setError(undefined)
                setResponse(callerResponse.response)
            } else {
                // For a 4xx or 5xx response this is where we provide any standardized logic for
                // error handling. For example:
                //
                // - On a 401 response, we need to force-logout the user so they can refresh their
                //   SAMS access token.
                // - On a 500 response, perhaps replace the current UI with a full-page error. e.g.
                //   http://github.com/500 or http://github.com/501
                setData(undefined)
                setError(new Error(`unexpected status code: ${callerResponse.response.status}`))
                setResponse(callerResponse.response)

                // Provide a clearer message. A 401 typically comes from the user's SAMS credentials
                // having expired on the backend.
                if (callerResponse.response.status === 401) {
                    setError(new Error('Please log out and log back in.'))
                }
            }
        } catch (error) {
            setData(undefined)
            setError(error)
            setResponse(undefined)
        } finally {
            setLoading(false)
        }
    }, [call, caller])

    useEffect(() => {
        void callApi()
    }, [callApi])

    return { loading, error, data, response, refetch: callApi }
}
