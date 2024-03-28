import React, { useEffect, useState } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../../auth'

import { SurveyToastContent } from './SurveyToastContent'

interface SurveyToastTriggerProps extends TelemetryV2Props {
    /**
     * For Storybook only
     */
    forceVisible?: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const SurveyToastTrigger: React.FunctionComponent<React.PropsWithChildren<SurveyToastTriggerProps>> = ({
    forceVisible,
    authenticatedUser,
    telemetryRecorder,
}) => {
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

    const visible = forceVisible || shouldShow

    if (!visible) {
        return null
    }

    return (
        <SurveyToastContent
            authenticatedUser={authenticatedUser}
            shouldTemporarilyDismiss={() => setTemporarilyDismissed(true)}
            shouldPermanentlyDismiss={() => setPermanentlyDismissed(true)}
            hideToast={() => setShouldShow(false)}
            telemetryRecorder={telemetryRecorder}
        />
    )
}
