import React, { useEffect, useState } from 'react'

import { Checkbox } from '@sourcegraph/wildcard'

import { useTemporarySetting } from '../settings/temporary/useTemporarySetting'
import { eventLogger } from '../tracking/eventLogger'

import { SurveyRatingRadio } from './SurveyRatingRadio'
import { Toast } from './Toast'
import { getDaysActiveCount } from './util'

interface SurveyToastProps {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
}

export const SurveyToast: React.FunctionComponent<SurveyToastProps> = ({ forceVisible }) => {
    const [toastDismissal, setToastDismissal] = useTemporarySetting('survey.toastDismissal')
    const [shouldPermanentlyDismissToast, setShouldPermanentlyDismissToast] = useState(false)
    const daysActive = getDaysActiveCount()

    console.log('toastDismissal', toastDismissal)

    const visible = forceVisible || (!toastDismissal && daysActive % 30 === 3)

    useEffect(() => {
        if (visible) {
            eventLogger.log('SurveyReminderViewed')
        }
    }, [visible])

    useEffect(() => {
        if (toastDismissal && toastDismissal !== 'permanent' && daysActive % 30 === 0) {
            setToastDismissal(() => undefined)
        }
    }, [daysActive, setToastDismissal, toastDismissal])

    const handleDismiss = (): void => setToastDismissal(shouldPermanentlyDismissToast ? 'permanent' : 'temporary')

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
                    checked={shouldPermanentlyDismissToast}
                    onChange={event => setShouldPermanentlyDismissToast(event.target.checked)}
                />
            }
            onDismiss={handleDismiss}
        />
    )
}
