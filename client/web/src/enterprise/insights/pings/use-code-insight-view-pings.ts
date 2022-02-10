import { useCallback, useRef } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { InsightType } from '../core/types'

interface UseCodeInsightViewPingsInput extends TelemetryProps {
    /**
     * View tracking type is used to send a proper pings event (InsightHover, InsightDataPointClick)
     * with view type as a tracking variable.
     */
    viewType: InsightType

    /**
     * Prefix for the all code insight view pings (InsightHover, InsightDataPointClick) It's used to be able
     * to tune pings event names in order to reuse logic but send specific for consumer pings with special names.
     */
    pingEventPrefix?: string
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
    const { viewType, pingEventPrefix = '', telemetryService } = props
    const timeoutID = useRef<number>()

    const trackMouseEnter = useCallback(() => {
        // Set timer to increase confidence that the user meant to interact with the
        // view, as opposed to accidentally moving past it. If the mouse leaves
        // the view quickly, clear the timeout for logging the event
        timeoutID.current = window.setTimeout(() => {
            telemetryService.log(`${pingEventPrefix}InsightHover`, { insightType: viewType }, { insightType: viewType })
        }, 500)
    }, [viewType, pingEventPrefix, telemetryService])

    const trackMouseLeave = useCallback(() => {
        window.clearTimeout(timeoutID.current)
    }, [])

    const trackDatumClicks = useCallback(() => {
        telemetryService.log(
            `${pingEventPrefix}InsightDataPointClick`,
            { insightType: viewType },
            { insightType: viewType }
        )
    }, [viewType, pingEventPrefix, telemetryService])

    return {
        trackDatumClicks,
        trackMouseEnter,
        trackMouseLeave,
    }
}
