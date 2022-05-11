import React, { useState } from 'react'

import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { useMutation, gql } from '@sourcegraph/http-client'
import { Button, LoadingSpinner, TextArea } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { SubmitSurveyResult, SubmitSurveyVariables } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { SurveyRatingRadio } from './SurveyRatingRadio'

import styles from './SurveyPage.module.scss'

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

export const SurveyForm: React.FunctionComponent<React.PropsWithChildren<SurveyFormProps>> = ({
    authenticatedUser,
    score,
}) => {
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
                <TextArea
                    id="survey-form-score-reason"
                    onChange={event => setReason(event.target.value)}
                    value={reason}
                    disabled={response.loading}
                    autoFocus={true}
                    label={
                        <span className={styles.label}>
                            What is the most important reason for the score you gave Sourcegraph?
                        </span>
                    }
                />
            </div>
            <div className="form-group">
                <TextArea
                    id="survey-form-better-product"
                    onChange={event => setBetterProduct(event.target.value)}
                    value={betterProduct}
                    disabled={response.loading}
                    label={<span className={styles.label}>What could Sourcegraph do to provide a better product?</span>}
                />
            </div>
            <div className="form-group">
                <Button display="block" variant="primary" type="submit" disabled={response.loading}>
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
