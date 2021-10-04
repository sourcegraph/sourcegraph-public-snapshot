import EmoticonIcon from 'mdi-react/EmoticonIcon'
import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '../auth'
import { eventLogger } from '../tracking/eventLogger'

import { SurveyCTA } from './SurveyCta'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

interface SurveyToastProps {
    authenticatedUser: AuthenticatedUser | null
}

export const SurveyToast: React.FunctionComponent<SurveyToastProps> = () => {
    const [visible, setVisible] = useState(
        localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && daysActiveCount % 30 === 3
    )

    useEffect(() => {
        if (visible) {
            eventLogger.log('SurveyReminderViewed')
        } else if (daysActiveCount % 30 === 0) {
            // Reset toast dismissal 3 days before it will be shown
            localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'false')
        }
    }, [visible])

    const handleDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        // TODO: Check weird condition where this is immediately set back to false if 3 days before shown again
        setVisible(false)
    }

    if (!visible) {
        return null
    }

    return (
        <Toast
            icon={<EmoticonIcon className="icon-inline" />}
            title="Tell us what you think"
            subtitle="How likely is it that you would recommend Sourcegraph to a friend?"
            cta={<SurveyCTA onChange={handleDismiss} openSurveyInNewTab={true} />}
            onDismiss={handleDismiss}
        />
    )
}
