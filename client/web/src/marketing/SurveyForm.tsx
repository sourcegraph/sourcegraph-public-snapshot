import React, { useState } from 'react'

import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { useMutation, gql } from '@sourcegraph/http-client'
import { Button, LoadingSpinner, Typography, Text, Input } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { SubmitSurveyResult, SubmitSurveyVariables } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { SurveyRatingRadio } from './SurveyRatingRadio'
import { SurveyUseCaseForm } from './SurveyUseCaseForm'

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
    const [email, setEmail] = useState('')
    const [validationError, setValidationError] = useState<Error | null>(null)
    const [_useCases, setUseCases] = useState<string[]>([])
    const [otherUseCase, setOtherUseCase] = useState<string>('')
    const [moreSharedInfo, setMoreSharedInfo] = useState<string>('')

    const [submitSurvey, response] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(SUBMIT_SURVEY, {
        onCompleted: () => {
            history.push({
                pathname: '/survey/thanks',
                state: {
                    // Mutation is only submitted when score is defined
                    score: score!,
                    feedback: moreSharedInfo,
                },
            })
        },
    })

    const handleScoreChange = (newScore: number): void => {
        if (validationError) {
            setValidationError(null)
        }

        history.push(`/survey/${newScore}`)
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
                    reason: moreSharedInfo,
                    better: otherUseCase,

                    // // TODO: Update api to recieve params
                    // useCases,
                    // otherUseCase
                    // moreSharedInfo
                },
            },
        })
    }

    const error = validationError || response.error

    return (
        <Form className={styles.surveyForm} onSubmit={handleSubmit}>
            {error && <Text className={styles.error}>{error.message}</Text>}
            {/* Label is associated with control through aria-labelledby */}
            {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
            <Typography.Label id="survey-form-scores" className={styles.label}>
                How likely is it that you would recommend Sourcegraph to a friend?
            </Typography.Label>
            <SurveyRatingRadio ariaLabelledby="survey-form-scores" onChange={handleScoreChange} score={score} />
            {!authenticatedUser && (
                <div className="form-group">
                    <Input
                        placeholder="Email"
                        onChange={event => setEmail(event.target.value)}
                        value={email}
                        disabled={response.loading}
                    />
                </div>
            )}
            <SurveyUseCaseForm
                className="my-2"
                formLabelClassName={styles.label}
                title="You are using sourcegraph to..."
                onChangeUseCases={value => setUseCases(value)}
                otherUseCase={otherUseCase}
                onChangeOtherUseCase={others => setOtherUseCase(others)}
                moreSharedInfo={moreSharedInfo}
                onChangeMoreShareInfo={moreInfo => setMoreSharedInfo(moreInfo)}
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
