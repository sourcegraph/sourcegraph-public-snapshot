import React, { useEffect, useState } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { eventLogger } from '../tracking/eventLogger'

import { ThankYou } from './ThankYou'
import { UseCaseForm } from './UseCaseForm'
import { UserRatingForm } from './UserRatingForm'

interface SurveyToastProps {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
}

export interface userFeedbackProps {
    score: number | string
    useCases: string[]
    moreSharedInfo: string
    otherUseCase: string
}

enum ToastSteps {
    rate = 1,
    useCase = 2,
    thankYou = 3,
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
    const [activeStep, setActiveStep] = useState(ToastSteps.rate)
    const [userFeedback, setUserFeedback] = useState<userFeedbackProps>({
        score: '',
        useCases: [],
        otherUseCase: '',
        moreSharedInfo: '',
    })

    const handleContinue = (): void => {
        if (userFeedback.score !== '') {
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

    const handleDismiss = (): void => {
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

    const handleSubmit = (): void => {
        // TODO: Send userFeedback to backend
        handleDismiss()
    }

    if (!visible) {
        return null
    }

    switch (activeStep) {
        case ToastSteps.useCase:
            return (
                <UseCaseForm
                    handleDone={props => {
                        setUserFeedback(current => ({ ...current, ...props }))
                        handleContinue()
                    }}
                    handleDismiss={handleDismiss}
                />
            )
        case ToastSteps.thankYou:
            return (
                <ThankYou
                    handleSubmit={() => {
                        handleSubmit()
                    }}
                />
            )
        case ToastSteps.rate:
        default:
            return (
                <UserRatingForm
                    onChange={score => setUserFeedback(current => ({ ...current, score }))}
                    handleDismiss={handleDismiss}
                    handleContinue={handleContinue}
                    toggleShouldPermanentlyDismiss={toggleShouldPermanentlyDismiss}
                    shouldPermanentlyDismiss={shouldPermanentlyDismiss}
                    toggleErrorMessage={toggleErrorMessage}
                />
            )
    }
}
