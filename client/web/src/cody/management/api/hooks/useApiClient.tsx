import { useEffect, useState, useContext } from 'react'

import { Call, Client } from '../client'
import { CodyProApiClientContext } from '../components/CodyProApiClient'

// useApiClient returns the Cody Pro API client. This isn't useful on its own, and
// needs to be paired with the `userApiCaller` hook.
export function useApiClient(): Client {
    const { client } = useContext(CodyProApiClientContext)
    return client
}

export interface ReactFriendlyApiResponse<T> {
    loading: boolean
    error?: Error
    data?: T
    response?: Response
}

// useApiCaller will issue a REST API call to the backend, returning the
// loading status and the response JSON object and/or error object as React
// state.
export function useApiCaller<Resp>(call: Call<Resp>): ReactFriendlyApiResponse<Resp> {
    const { caller } = useContext(CodyProApiClientContext)

    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<any>(undefined)
    const [data, setData] = useState<any>(undefined)
    const [response, setResponse] = useState<any>(undefined)

    useEffect(() => {
        ;(async () => {
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
                    setError(Error(`unexpected status code: ${callerResponse.response.status}`))
                    setResponse(callerResponse.response)

                    // Provide a clearer message. A 401 typically comes from the user's SAMS credentials
                    // having expired on the backend.
                    if (callerResponse.response.status === 401) {
                        setError(Error('Please log out and log back in.'))
                    }
                }
                setLoading(false)
            } catch (err) {
                setData(undefined)
                setError(err)
                setResponse(undefined)
                setLoading(false)
            }
        })()
    }, [caller])

    return { loading, error, data, response }
}
