import React, { useState } from 'react'
import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { useMutation } from '@sourcegraph/shared/src/graphql/apollo'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { SubmitSurveyResult, SubmitSurveyVariables } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import styles from './SurveyPage.module.scss'
import { SurveyRatingRadio } from './SurveyRatingRadio'

interface SurveyFormProps {
    authenticatedUser: AuthenticatedUser | null
    score?: number
}

export const SUBMIT_SURVEY = gql`
    mutation SubmitSurvey($input: SurveySubmissionInput!) {
        submitSurvey(input: $input) {
            alwaysNil
        }
    }
`

/**
 * Form Data that is persisted to `location` and retrieved in other components.
 */
export interface SurveyFormLocationState {
    score: number
    feedback: string
}

export const SurveyForm: React.FunctionComponent<SurveyFormProps> = ({ authenticatedUser, score }) => {
    const history = useHistory<SurveyFormLocationState>()
    const [reason, setReason] = useState('')
    const [betterProduct, setBetterProduct] = useState('')
    const [email, setEmail] = useState('')
    const [validationError, setValidationError] = useState<Error | null>(null)

    const [submitSurvey, response] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(SUBMIT_SURVEY, {
        onCompleted: () => {
            history.push({
                pathname: '/survey/thanks',
                state: {
                    // Mutation is only submitted when score is defined
                    score: score!,
                    feedback: reason,
                },
            })
        },
    })

    const handleScoreChange = (): void => {
        if (validationError) {
            setValidationError(null)
        }
    }

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>): Promise<void> => {
        event.preventDefault()

        if (score === undefined) {
            setValidationError(new Error('Please select a score'))
            return
        }

        eventLogger.log('SurveySubmitted')

        await submitSurvey({
            variables: {
                input: {
                    score,
                    email,
                    reason,
                    better: betterProduct,
                },
            },
        })
    }

    const error = validationError || response.error

    return (
        <Form className={styles.surveyForm} onSubmit={handleSubmit}>
            {error && <p className={styles.error}>{error.message}</p>}
            {/* Label is associated with control through aria-labelledby */}
            {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
            <label id="survey-form-scores" className={styles.label}>
                How likely is it that you would recommend Sourcegraph to a friend?
            </label>
            <SurveyRatingRadio ariaLabelledby="survey-form-scores" onChange={handleScoreChange} score={score} />
            {!authenticatedUser && (
                <div className="form-group">
                    <input
                        className="form-control"
                        type="text"
                        placeholder="Email"
                        onChange={event => setEmail(event.target.value)}
                        value={email}
                        disabled={response.loading}
                    />
                </div>
            )}
            <div className="form-group">
                <label className={styles.label} htmlFor="survey-form-score-reason">
                    What is the most important reason for the score you gave Sourcegraph?
                </label>
                <textarea
                    id="survey-form-score-reason"
                    className="form-control"
                    onChange={event => setReason(event.target.value)}
                    value={reason}
                    disabled={response.loading}
                    autoFocus={true}
                />
            </div>
            <div className="form-group">
                <label className={styles.label} htmlFor="survey-form-better-product">
                    What could Sourcegraph do to provide a better product?
                </label>
                <textarea
                    id="survey-form-better-product"
                    className="form-control"
                    onChange={event => setBetterProduct(event.target.value)}
                    value={betterProduct}
                    disabled={response.loading}
                />
            </div>
            <div className="form-group">
                <Button className="btn-block" variant="primary" type="submit" disabled={response.loading}>
                    Submit
                </Button>
            </div>
            {response.loading && (
                <div className={styles.loader}>
                    <LoadingSpinner />
                </div>
            )}
            <div>
                <small>
                    Your response to this survey will be sent to Sourcegraph, and will be visible to your Sourcegraph
                    site admins.
                </small>
            </div>
        </Form>
    )
}
