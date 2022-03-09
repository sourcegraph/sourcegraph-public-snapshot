import { useCallback } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useGettingStartedTourState } from '../stores/gettingStartedTourState'

export function useGettingStartedTourLogEvent(
    telemetryService: TelemetryProps['telemetryService']
): (eventName: string) => void {
    const language = useGettingStartedTourState(useCallback(state => state.language, []))

    return useCallback(
        (eventName: string) => {
            const args = { language }
            telemetryService.log(eventName, args, args)
        },
        [language, telemetryService]
    )
}
