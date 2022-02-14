import { useCallback } from 'react'

import { gql, useMutation } from '@sourcegraph/http-client'
import { FeedbackPromptSubmitEventHandler } from '@sourcegraph/wildcard'

import { SubmitHappinessFeedbackResult, SubmitHappinessFeedbackVariables } from '../graphql-operations'
import { LayoutRouteProps } from '../routes'

import { useRoutesMatch } from './useRoutesMatch'

interface HandleSubmitFeedbackState {
    handleSubmitFeedback: FeedbackPromptSubmitEventHandler
}

export const useHandleSubmitFeedback = (
    routes?: readonly LayoutRouteProps<{}>[] | any[],
    textPrefix = ''
): HandleSubmitFeedbackState => {
    const match = useRoutesMatch(routes)
    const SUBMIT_HAPPINESS_FEEDBACK_QUERY = gql`
        mutation SubmitHappinessFeedback($input: HappinessFeedbackSubmissionInput!) {
            submitHappinessFeedback(input: $input) {
                alwaysNil
            }
        }
    `

    const [submitFeedback] = useMutation<SubmitHappinessFeedbackResult, SubmitHappinessFeedbackVariables>(
        SUBMIT_HAPPINESS_FEEDBACK_QUERY
    )

    const handleSubmitFeedback = useCallback(
        async (text: string, rating: number) => {
            const { data, errors } = await submitFeedback({
                variables: {
                    input: { score: rating, feedback: `${textPrefix}${text}`, currentPath: match },
                },
            })

            return {
                errorMessage: errors?.map(error => error.message).join(', '),
                isHappinessFeedback: !!data?.submitHappinessFeedback,
            }
        },
        [match, submitFeedback, textPrefix]
    )

    return {
        handleSubmitFeedback,
    }
}
