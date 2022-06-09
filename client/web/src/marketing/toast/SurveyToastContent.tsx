import React, { useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'

import { AuthenticatedUser } from '../../auth'
import { SubmitSurveyResult, SubmitSurveyVariables, SurveyUseCase } from '../../graphql-operations'
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

interface UserFeedbackProps {
    score: number
    useCases: SurveyUseCase[]
    additionalInformation: string
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
    const [userFeedback, setUserFeedback] = useState<UserFeedbackProps>({
        score: -1,
        useCases: [],
        otherUseCase: '',
        additionalInformation: '',
        email: '',
    })

    useEffect(() => {
        eventLogger.log('SurveyReminderViewed')
    }, [])

    const setFutureVisibility = (): void => {
        if (togglePermanentlyDismiss) {
            shouldPermanentlyDismiss()
        } else {
            shouldTemporarilyDismiss()
        }
    }

    const [submitSurvey, { loading: isSubmitting }] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(
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
                    onChange={formState => setUserFeedback(current => ({ ...current, ...formState }))}
                    onDismiss={handleDismiss}
                    onDone={handleUseCaseDone}
                    authenticatedUser={authenticatedUser}
                />
            )
        case ToastSteps.thankYou:
            return <SurveySuccessToast onDismiss={handleDismiss} />
        default:
            throw new Error('Invalid survey step!')
    }
}
