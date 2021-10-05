import React, { useEffect, useState } from 'react'

import { Checkbox } from '@sourcegraph/wildcard'

import { TemporarySettingsSchema } from '../settings/temporary/TemporarySettings'
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
const shouldShowToast = (dismissed?: TemporarySettingsSchema['survey.toastDismissed']): boolean =>
    dismissed !== undefined && dismissed === false && getDaysActiveCount() % 30 === 3

/**
 * TODO: Flash of content
 */
export const SurveyToast: React.FunctionComponent<SurveyToastProps> = ({ forceVisible }) => {
    const [shouldPermanentlyDismiss, setShouldPermanentlyDismiss] = useState(false)
    const [toastDismissed, setToastDismissed] = useTemporarySetting('survey.toastDismissed', false)
    const visible = forceVisible || shouldShowToast(toastDismissed)

    useEffect(() => {
        if (visible) {
            eventLogger.log('SurveyReminderViewed')
        }
    }, [visible])

    useEffect(() => {
        // Reset 3 days before something
        if (toastDismissed === 'temporarily' && getDaysActiveCount() % 30 === 0) {
            setToastDismissed(false)
        }
    }, [toastDismissed, setToastDismissed])

    const handleDismiss = (): void => {
        setToastDismissed(shouldPermanentlyDismiss ? 'permanently' : 'temporarily')
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
                    onChange={event => setShouldPermanentlyDismiss(event.target.checked)}
                />
            }
            onDismiss={handleDismiss}
        />
    )
}
