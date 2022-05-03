import React, { useEffect, useState } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { eventLogger } from '../tracking/eventLogger'

import { SurveySuccess } from './SurveySuccess'
import { SurveyUseCaseForm } from './SurveyUseCaseForm'
import { SurveyUserRatingForm } from './SurveyUserRatingForm'

interface SurveyToastProps {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
}

interface UserFeedbackProps {
    score: number
    useCases: string[]
    moreSharedInfo: string
    otherUseCase: string
}

enum ToastSteps {
    rate = 1,
    useCase = 2,
    thankYou = 3,
}

const handleSubmit = (): void => {
    // TODO: Send <userFeedback> to backend
    // Moved out of component score temporarily.
}

export const SurveyToast: React.FunctionComponent<SurveyToastProps> = ({ forceVisible }) => {
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
        moreSharedInfo: '',
    })

    const handleContinue = (): void => {
        if (userFeedback.score !== -1) {
            setActiveStep(current => current + 1)
        } else {
            setToggleErrorMessage(true)
        }
    }

    const toggleShouldPermanentlyDismiss = (value: boolean): void => setShouldPermanentlyDismiss(value)

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
        if (activeStep === ToastSteps.useCase) {
            // TODO: Send userFeedback to backend.
            handleSubmit()
        }

        if (shouldPermanentlyDismiss) {
            setPermanentlyDismissed(shouldPermanentlyDismiss)
        } else {
            setTemporarilyDismissed(true)
        }
    }

    if (!visible) {
        return null
    }

    switch (activeStep) {
        case ToastSteps.useCase:
            return (
                <SurveyUseCaseForm
                    handleDone={formState => {
                        setUserFeedback(current => ({ ...current, ...formState }))
                        handleSubmit()
                        handleContinue()
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
                    toggleShouldPermanentlyDismiss={toggleShouldPermanentlyDismiss}
                    shouldPermanentlyDismiss={shouldPermanentlyDismiss}
                    toggleErrorMessage={toggleErrorMessage}
                />
            )
        default:
            return null
    }
}
