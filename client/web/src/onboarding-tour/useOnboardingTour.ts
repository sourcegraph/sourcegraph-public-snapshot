import { useCallback } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useOnboardingTourState } from '../stores/onboardingTourState'

export function useLogTourEvent(telemetryService: TelemetryProps['telemetryService']): (eventName: string) => void {
    const language = useOnboardingTourState(useCallback(state => state.language, []))

    return useCallback(
        (eventName: string) => {
            const args = { language }
            telemetryService.log(eventName, args, args)
        },
        [language, telemetryService]
    )
}
