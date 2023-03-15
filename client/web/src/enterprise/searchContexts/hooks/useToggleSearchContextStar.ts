import { useCallback, useEffect, useState } from 'react'

import { isErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'

export function useToggleSearchContextStar(
    initialStarred: boolean,
    searchContextId: string | undefined,
    userId: string | undefined
): { starred: boolean; toggleStar: () => Promise<void> } {
    const [starred, setStarred] = useState(initialStarred)
    useEffect(() => {
        setStarred(initialStarred)
    }, [initialStarred])

    const [createStarMutation] = useMutation(gql`
        mutation createSearchContextStar($searchContextID: ID!, $userID: ID!) {
            createSearchContextStar(searchContextID: $searchContextID, userID: $userID) {
                alwaysNil
            }
        }
    `)

    const [deleteStarMutation] = useMutation(gql`
        mutation deleteSearchContextStar($searchContextID: ID!, $userID: ID!) {
            deleteSearchContextStar(searchContextID: $searchContextID, userID: $userID) {
                alwaysNil
            }
        }
    `)

    const toggleStar = useCallback(async () => {
        const errorPrefix = `Failed to ${starred ? 'unstar' : 'star'} search context`

        // Cannot star if user is not authenticated
        if (!userId) {
            throw new Error(`${errorPrefix}: You are not signed in. Please sign in and try again.`)
        }

        const previousStarred = starred

        // Optimistically update the state
        setStarred(!previousStarred)

        const promise = previousStarred
            ? deleteStarMutation({ variables: { searchContextID: searchContextId, userID: userId } })
            : createStarMutation({ variables: { searchContextID: searchContextId, userID: userId } })

        try {
            await promise
        } catch (error) {
            // If the mutation fails, revert the optimistic update
            setStarred(previousStarred)

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
    }, [createStarMutation, deleteStarMutation, searchContextId, starred, userId])

    return { starred, toggleStar }
}
