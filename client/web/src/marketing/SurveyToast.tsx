import React, { useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { SubmitSurveyResult, SubmitSurveyVariables, SurveyUseCase } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { SurveySuccess } from './SurveySuccess'
import { SurveyUseCaseToast } from './SurveyUseCaseToast'
import { SurveyUserRatingForm } from './SurveyUserRatingForm'

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
}

interface UserFeedbackProps {
    score: number
    useCases: SurveyUseCase[]
    additionalInformation: string
    otherUseCase: string
}

enum ToastSteps {
    rate = 1,
    useCase = 2,
    thankYou = 3,
}

export const SurveyToast: React.FunctionComponent<React.PropsWithChildren<SurveyToastProps>> = ({ forceVisible }) => {
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

    const [toggleErrorMessage, setToggleErrorMessage] = useState<boolean>(false)
    const [activeStep, setActiveStep] = useState<ToastSteps>(ToastSteps.rate)
    const [userFeedback, setUserFeedback] = useState<UserFeedbackProps>({
        score: -1,
        useCases: [],
        otherUseCase: '',
        additionalInformation: '',
    })

    const [submitSurvey] = useMutation<SubmitSurveyResult, SubmitSurveyVariables>(SUBMIT_SURVEY)

    const handleSubmit = async (): Promise<void> => {
        await submitSurvey({
            variables: {
                input: {
                    ...userFeedback,
                    email: '',
                },
            },
        })
    }

    const handleContinue = (): void => {
        if (userFeedback.score !== -1) {
            setActiveStep(current => current + 1)
        } else {
            setToggleErrorMessage(true)
        }
    }

    /**
     * We show a toast notification if:
     * 1. User has not recently dismissed the notification
     * 2. User has not permanently dismissed the notification
     * 3. User has been active for exactly 3 days OR it has been 30 days since they were last shown the notification
     */
    const shouldShow =
        !loadingTemporarySettings && !temporarilyDismissed && !permanentlyDismissed && daysActiveCount % 30 === 3

    const visible = forceVisible || shouldShow

    useEffect(() => {
        if (visible) {
            eventLogger.log('SurveyReminderViewed')
        }
    }, [visible])

    useEffect(() => {
        if (!loadingTemporarySettings && daysActiveCount % 30 === 0) {
            // Reset toast dismissal 3 days before it will be shown
            setTemporarilyDismissed(false)
        }
    }, [loadingTemporarySettings, daysActiveCount, setTemporarilyDismissed])

    const onDismiss = (): void => {
        if (userFeedback.score !== -1 && activeStep !== ToastSteps.thankYou) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            handleSubmit()
        }

        if (shouldPermanentlyDismiss) {
            setPermanentlyDismissed(shouldPermanentlyDismiss)
        } else {
            setTemporarilyDismissed(true)
        }
    }

    const handleUseCaseDone = async (): Promise<void> => {
        await handleSubmit()
        handleContinue()
    }

    if (!visible) {
        return null
    }

    switch (activeStep) {
        case ToastSteps.useCase:
            return (
                <SurveyUseCaseToast
                    onDone={handleUseCaseDone}
                    onChange={formState => {
                        setUserFeedback(current => ({ ...current, ...formState }))
                    }}
                    onDismiss={onDismiss}
                />
            )
        case ToastSteps.thankYou:
            return <SurveySuccess onDismiss={onDismiss} />
        case ToastSteps.rate:
            return (
                <SurveyUserRatingForm
                    onChange={score => setUserFeedback(current => ({ ...current, score }))}
                    onDismiss={onDismiss}
                    onContinue={handleContinue}
                    toggleShouldPermanentlyDismiss={setShouldPermanentlyDismiss}
                    shouldPermanentlyDismiss={shouldPermanentlyDismiss}
                    toggleErrorMessage={toggleErrorMessage}
                />
            )
        default:
            throw new Error('Invalid survey step!')
    }
}
