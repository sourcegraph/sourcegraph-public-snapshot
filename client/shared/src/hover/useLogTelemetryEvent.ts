import { useEffect, useRef } from 'react'

import { isEqual } from 'lodash'

import { isErrorLike } from '@sourcegraph/common'

import type { HoverOverlayProps } from './HoverOverlay'

const isEmptyHover = ({
    hoveredToken,
    hoverOrError,
    actionsOrError,
}: Pick<HoverOverlayProps, 'hoveredToken' | 'hoverOrError' | 'actionsOrError'>): boolean =>
    !hoveredToken ||
    ((!hoverOrError || hoverOrError === 'loading' || isErrorLike(hoverOrError)) &&
        (!actionsOrError || actionsOrError === 'loading' || isErrorLike(actionsOrError)))

// Log telemetry event on mount and once per new hover position
export function useLogTelemetryEvent(props: HoverOverlayProps): void {
    const { telemetryService, telemetryRecorder, hoveredToken } = props

    const previousPropsReference = useRef(props)
    const logTelemetryEvent = (): void => {
        telemetryService.log('hover')
        telemetryRecorder.recordEvent('blob', 'hover')
    }

    // Log a telemetry event on component mount once, so we don't care about dependency updates.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => logTelemetryEvent(), [])

    // Log a telemetry event for this hover being displayed,
    // but only do it once per position and when it is non-empty.
    if (
        !isEmptyHover(props) &&
        (!isEqual(hoveredToken, previousPropsReference.current.hoveredToken) ||
            isEmptyHover(previousPropsReference.current))
    ) {
        logTelemetryEvent()
    }

    // Update previous props ref after we used it to determine if a telemetry event log is needed.
    previousPropsReference.current = props
}
