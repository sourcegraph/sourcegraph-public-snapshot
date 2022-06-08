import React, { useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

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

interface SurveyToastProps {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
    authenticatedUser: AuthenticatedUser | null
}

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

export const SurveyToast: React.FunctionComponent<React.PropsWithChildren<SurveyToastProps>> = ({
    forceVisible,
    authenticatedUser,
}) => {
    const [shouldPermanentlyDismiss, setShouldPermanentlyDismiss] = useState(false)
    const [temporarilyDismissed, setTemporarilyDismissed] = useTemporarySetting(
        'npsSurvey.hasTemporarilyDismissed',
        false
    )
    const [permanentlyDismissed, setPermanentlyDismissed] = useTemporarySetting(
        'npsSurvey.hasPermanentlyDismissed',
        false
    )
    const [daysActiveCount] = useTemporarySetting('user.daysActiveCount', 0)
    const loadingTemporarySettings =
        temporarilyDismissed === undefined || permanentlyDismissed === undefined || daysActiveCount === undefined

    const [shouldShow, setShouldShow] = useState(false)
    const [toggleErrorMessage, setToggleErrorMessage] = useState<boolean>(false)
    const [activeStep, setActiveStep] = useState<ToastSteps>(ToastSteps.rate)
    const [userFeedback, setUserFeedback] = useState<UserFeedbackProps>({
        score: -1,
        useCases: [],
        otherUseCase: '',
        additionalInformation: '',
        email: '',
    })

    /**
     * We set dismissal state when either:
     * 1. User clicks the dismiss button
     * 2. User submits the survey
     */
    const updateDismissalState = (): void => {
        if (shouldPermanentlyDismiss) {
            setPermanentlyDismissed(shouldPermanentlyDismiss)
        } else {
            setTemporarilyDismissed(true)
        }
    }

    const [submitSurvey, { loading: isSubmitting }] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(
        SUBMIT_SURVEY,
        {
            onCompleted: updateDismissalState,
        }
    )

    useEffect(() => {
        if (!loadingTemporarySettings) {
            /**
             * We show a toast notification if:
             * 1. User has not recently dismissed the notification
             * 2. User has not permanently dismissed the notification
             * 3. User has been active for exactly 3 days OR it has been 30 days since they were last shown the notification
             */
            setShouldShow(!temporarilyDismissed && !permanentlyDismissed && daysActiveCount % 30 === 3)
        }

        /**
         * We only use the initial temporary settings to ensure we have better control over when the toast is shown.
         * E.g. we want to always update temporary settings on submit, but show a thank you screen before dismissal.
         */
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [loadingTemporarySettings])

    useEffect(() => {
        if (!loadingTemporarySettings && daysActiveCount % 30 === 0) {
            // Reset toast dismissal 3 days before it will be shown
            setTemporarilyDismissed(false)
        }
    }, [loadingTemporarySettings, daysActiveCount, setTemporarilyDismissed])

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
            // No need to submit, but we want to ensure user isn't bothered by the toast again
            updateDismissalState()
        }

        // Hide the toast
        setShouldShow(false)
    }

    const handleUseCaseDone = async (): Promise<void> => {
        await submitSurvey({ variables: { input: userFeedback } })
        handleContinue()
    }

    const visible = forceVisible || shouldShow

    useEffect(() => {
        if (visible) {
            eventLogger.log('SurveyReminderViewed')
        }
    }, [visible])

    if (!visible) {
        return null
    }

    switch (activeStep) {
        case ToastSteps.rate:
            return (
                <SurveyUserRatingToast
                    onChange={score => setUserFeedback(current => ({ ...current, score }))}
                    onDismiss={handleDismiss}
                    onContinue={handleContinue}
                    toggleShouldPermanentlyDismiss={setShouldPermanentlyDismiss}
                    shouldPermanentlyDismiss={shouldPermanentlyDismiss}
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
