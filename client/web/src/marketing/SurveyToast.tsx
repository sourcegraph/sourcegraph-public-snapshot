import React, { useEffect, useState } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Checkbox } from '@sourcegraph/wildcard'

import { eventLogger } from '../tracking/eventLogger'

import { SurveyRatingRadio } from './SurveyRatingRadio'
import { Toast } from './Toast'

interface SurveyToastProps {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
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
        if (shouldPermanentlyDismiss) {
            setPermanentlyDismissed(shouldPermanentlyDismiss)
        } else {
            setTemporarilyDismissed(true)
        }
    }

    if (!visible) {
        return null
    }

    return (
        <Toast
            title="Tell us what you think"
            subtitle={
                <span id="survey-toast-scores">How likely is it that you would recommend Sourcegraph to a friend?</span>
            }
            cta={
                <SurveyRatingRadio
                    onChange={handleDismiss}
                    openSurveyInNewTab={true}
                    ariaLabelledby="survey-toast-scores"
                />
            }
            footer={
                <Checkbox
                    id="survey-toast-refuse"
                    label="Don't show this again"
                    checked={shouldPermanentlyDismiss}
                    onChange={event => setShouldPermanentlyDismiss(event.target.checked)}
                />
            }
            onDismiss={handleDismiss}
        />
    )
}
