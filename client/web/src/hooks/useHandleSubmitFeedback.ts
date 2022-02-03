import { ApolloError, OperationVariables } from '@apollo/client'
import { useCallback } from 'react'

import { gql, useMutation } from '@sourcegraph/http-client'

import { SubmitHappinessFeedbackResult, SubmitHappinessFeedbackVariables } from '../graphql-operations'

interface useHandleSubmitFeedbackState {
    loading: boolean
    data?: SubmitHappinessFeedbackResult | null
    error?: ApolloError
    onSubmit: (text: string, rating: number) => Promise<OperationVariables | undefined>
}

interface useHandleSubmitFeedbackProps {
    routeMatch?: string
    textPrefix?: string
}

export const useHandleSubmitFeedback = ({
    textPrefix = '',
    routeMatch,
}: useHandleSubmitFeedbackProps): useHandleSubmitFeedbackState => {
    const SUBMIT_HAPPINESS_FEEDBACK_QUERY = gql`
        mutation SubmitHappinessFeedback($input: HappinessFeedbackSubmissionInput!) {
            submitHappinessFeedback(input: $input) {
                alwaysNil
            }
        }
    `

    const [submitFeedback, { loading, data, error }] = useMutation<
        SubmitHappinessFeedbackResult,
        SubmitHappinessFeedbackVariables
    >(SUBMIT_HAPPINESS_FEEDBACK_QUERY)

    const onSubmit = useCallback(
        (text: string, rating: number) =>
            submitFeedback({
                variables: {
                    input: { score: rating, feedback: `${textPrefix}${text}`, currentPath: routeMatch },
                },
            }),
        [routeMatch, submitFeedback, textPrefix]
    )

    return {
        loading,
        data,
        error,
        onSubmit,
    }
}
