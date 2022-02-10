import { useCallback, useRef } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { InsightType } from '../core/types'

interface UseCodeInsightViewPingsInput extends TelemetryProps {
    /**
     * View tracking type is used to send a proper pings event (InsightHover, InsightDataPointClick)
     * with view type as a tracking variable.
     */
    viewType: InsightType
}

interface PingHandlers {
    trackMouseEnter: () => void
    trackMouseLeave: () => void
    trackDatumClicks: () => void
}

/**
 * Shared logic for tracking insight related ping events on the insight card component.
 */
export function useCodeInsightViewPings(props: UseCodeInsightViewPingsInput): PingHandlers {
    const { viewType, telemetryService } = props
    const timeoutID = useRef<number>()

    const trackMouseEnter = useCallback(() => {
        // Set timer to increase confidence that the user meant to interact with the
        // view, as opposed to accidentally moving past it. If the mouse leaves
        // the view quickly, clear the timeout for logging the event
        timeoutID.current = window.setTimeout(() => {
            telemetryService.log('InsightHover', { insightType: viewType }, { insightType: viewType })
        }, 500)
    }, [viewType, telemetryService])

    const trackMouseLeave = useCallback(() => {
        window.clearTimeout(timeoutID.current)
    }, [])

    const trackDatumClicks = useCallback(() => {
        telemetryService.log('InsightDataPointClick', { insightType: viewType }, { insightType: viewType })
    }, [viewType, telemetryService])

    return {
        trackDatumClicks,
        trackMouseEnter,
        trackMouseLeave,
    }
}
