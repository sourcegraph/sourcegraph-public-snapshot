import React, { useState } from 'react'

import { useNavigate } from 'react-router-dom'

import { useMutation, gql } from '@sourcegraph/http-client'
import { Button, LoadingSpinner, Label, Text, Form } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import type { SubmitSurveyResult, SubmitSurveyVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { SurveyRatingRadio } from '../components/SurveyRatingRadio'
import { SurveyUseCaseForm } from '../components/SurveyUseCaseForm'

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
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [validationError, setValidationError] = useState<Error | null>(null)
    const [otherUseCase, setOtherUseCase] = useState<string>('')
    const [better, setBetter] = useState<string>('')

    const [submitSurvey, response] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(SUBMIT_SURVEY, {
        onCompleted: () => {
            navigate(
                {
                    pathname: '/survey/thanks',
                },
                {
                    state: {
                        // Mutation is only submitted when score is defined
                        score: score!,
                        feedback: better,
                    },
                }
            )
        },
    })

    const handleScoreChange = (newScore: number): void => {
        if (validationError) {
            setValidationError(null)
        }

        navigate(`/survey/${newScore}`)
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
                    email,
                    score,
                    otherUseCase,
                    better,
                },
            },
        })
    }

    const error = validationError || response.error

    return (
        <Form className={styles.surveyForm} onSubmit={handleSubmit}>
            {error && <Text className={styles.error}>{error.message}</Text>}
            {/* Label is associated with control through aria-labelledby */}
            {}
            <Label id="survey-form-scores" className={styles.label}>
                How likely is it that you would recommend Sourcegraph to a friend?
            </Label>
            <SurveyRatingRadio ariaLabelledby="survey-form-scores" onChange={handleScoreChange} score={score} />
            <SurveyUseCaseForm
                className="my-2"
                authenticatedUser={authenticatedUser}
                formLabelClassName={styles.label}
                otherUseCase={otherUseCase}
                onChangeOtherUseCase={setOtherUseCase}
                better={better}
                onChangeBetter={setBetter}
                email={email}
                onChangeEmail={setEmail}
            />
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
