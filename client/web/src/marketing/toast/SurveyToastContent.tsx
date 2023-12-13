import React, { useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'

import type { AuthenticatedUser } from '../../auth'
import type { SubmitSurveyResult, SubmitSurveyVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { SurveySuccessToast } from './SurveySuccessToast'
import { SurveyUseCaseToast } from './SurveyUseCaseToast'
import { SurveyUserRatingToast } from './SurveyUserRatingToast'

const SUBMIT_SURVEY = gql`
    mutation SubmitSurvey($input: SurveySubmissionInput!) {
        submitSurvey(input: $input) {
            alwaysNil
        }
    }
`

export interface TotalFeedbackState {
    score: number
    better: string
    otherUseCase: string
    email: string
}

enum ToastSteps {
    rate = 1,
    useCase = 2,
    thankYou = 3,
}

interface SurveyToastContentProps {
    authenticatedUser: AuthenticatedUser | null
    shouldTemporarilyDismiss: () => void
    shouldPermanentlyDismiss: () => void
    hideToast: () => void
}

export const SurveyToastContent: React.FunctionComponent<React.PropsWithChildren<SurveyToastContentProps>> = ({
    authenticatedUser,
    shouldTemporarilyDismiss,
    shouldPermanentlyDismiss,
    hideToast,
}) => {
    const [togglePermanentlyDismiss, setTogglePermanentlyDismiss] = useState(false)
    const [toggleErrorMessage, setToggleErrorMessage] = useState<boolean>(false)
    const [activeStep, setActiveStep] = useState<ToastSteps>(ToastSteps.rate)
    const [userFeedback, setUserFeedback] = useState<TotalFeedbackState>({
        score: -1,
        otherUseCase: '',
        better: '',
        email: '',
    })

    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('surveryReminder', 'viewed')
        eventLogger.log('SurveyReminderViewed')
    }, [])

    /**
     * We set dismissal state when either:
     * 1. User clicks the dismiss button
     * 2. User submits the survey
     *
     * It's important to separate the two actions, as it is possible the user might
     * exit the page after submitting the survey but BEFORE dismissing the button.
     */
    const setFutureVisibility = (): void => {
        if (togglePermanentlyDismiss) {
            shouldPermanentlyDismiss()
        } else {
            shouldTemporarilyDismiss()
        }
    }

    const [submitSurvey, { loading: isSubmitting, error }] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(
        SUBMIT_SURVEY,
        { onCompleted: setFutureVisibility }
    )

    const handleContinue = (): void => {
        if (userFeedback.score !== -1) {
            setActiveStep(current => current + 1)
        } else {
            setToggleErrorMessage(true)
        }
    }

    const handleDismiss = (): void => {
        /**
         * Ensures we still submit on dismiss when a user has (at least)
         * set a score AND they haven't already submitted.
         */
        if (userFeedback.score !== -1 && activeStep !== ToastSteps.thankYou) {
            // User is dismissing, we fire but don't wait for the response
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            submitSurvey({ variables: { input: userFeedback } })
        } else {
            // No need to submit, but we want to ensure user isn't bothered by the toast again before exiting
            setFutureVisibility()
        }

        hideToast()
    }

    const handleUseCaseDone = async (): Promise<void> => {
        await submitSurvey({ variables: { input: userFeedback } })
        handleContinue()
    }

    switch (activeStep) {
        case ToastSteps.rate:
            return (
                <SurveyUserRatingToast
                    score={userFeedback.score}
                    onChange={score => setUserFeedback(current => ({ ...current, score }))}
                    onDismiss={handleDismiss}
                    onContinue={handleContinue}
                    setToggledPermanentlyDismiss={setTogglePermanentlyDismiss}
                    toggleErrorMessage={toggleErrorMessage}
                />
            )
        case ToastSteps.useCase:
            return (
                <SurveyUseCaseToast
                    isSubmitting={isSubmitting}
                    otherUseCase={userFeedback.otherUseCase}
                    onChangeOtherUseCase={otherUseCase => setUserFeedback(current => ({ ...current, otherUseCase }))}
                    better={userFeedback.better}
                    onChangeBetter={better => setUserFeedback(current => ({ ...current, better }))}
                    email={userFeedback.email}
                    onChangeEmail={email => setUserFeedback(current => ({ ...current, email }))}
                    onDismiss={handleDismiss}
                    onDone={handleUseCaseDone}
                    error={error}
                    authenticatedUser={authenticatedUser}
                />
            )
        case ToastSteps.thankYou:
            return <SurveySuccessToast onDismiss={handleDismiss} />
        default:
            throw new Error('Invalid survey step!')
    }
}
