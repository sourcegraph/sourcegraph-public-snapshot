import { useCallback, useEffect, useState } from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'

export const SET_DEFAULT_SEARCH_CONTEXT_MUTATION = gql`
    mutation setDefaultSearchContext($searchContextID: ID!, $userID: ID!) {
        setDefaultSearchContext(searchContextID: $searchContextID, userID: $userID) {
            alwaysNil
        }
    }
`

export function useDefaultContext(initialDefaultSearchContextId: string | undefined): {
    defaultContext: string | undefined
    setAsDefault: (searchContextId?: string, userId?: string) => Promise<void>
} {
    const [defaultContext, setDefaultContext] = useState(initialDefaultSearchContextId)
    useEffect(() => {
        setDefaultContext(initialDefaultSearchContextId)
    }, [initialDefaultSearchContextId])

    const [setAsDefaultMutation] = useMutation(SET_DEFAULT_SEARCH_CONTEXT_MUTATION)

    const setAsDefault = useCallback(
        async (searchContextId?: string, userId?: string) => {
            const errorPrefix = 'Failed to set search context as default'

            // Cannot set as default if user is not authenticated
            if (!userId) {
                throw new Error(`${errorPrefix}: You are not signed in. Please sign in and try again.`)
            }

            // Optimistically update the state
            const previousDefaultContext = defaultContext
            setDefaultContext(searchContextId)

            try {
                await setAsDefaultMutation({
                    variables: { searchContextID: searchContextId, userID: userId },
                })
            } catch (error) {
                // If the mutation fails, revert the optimistic update
                setDefaultContext(previousDefaultContext)

                if (isErrorLike(error)) {
                    if (error.message.includes('Failed to fetch')) {
                        throw new Error(
                            `${errorPrefix}: Could not contact the server, please check your network connection and try again.`,
                            { cause: error }
                        )
                    } else {
                        throw new Error(`${errorPrefix}: ${error.message}`, { cause: error })
                    }
                } else {
                    throw new Error(`${errorPrefix}: An unknown error occurred.`, { cause: error })
                }
            }
        },
        [defaultContext, setAsDefaultMutation]
    )

    return { defaultContext, setAsDefault }
}
