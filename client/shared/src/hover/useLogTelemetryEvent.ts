import { useEffect, useRef } from 'react'

import { isEqual } from 'lodash'

import { isErrorLike } from '@sourcegraph/common'

import { EventName } from '../telemetry/telemetryServiceV2'

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
    const { telemetryService, hoveredToken } = props

    const previousPropsReference = useRef(props)
    const logTelemetryEvent = (): void => telemetryService.log('hover')

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

// Record telemetry event on mount and once per new hover position
export function useRecordTelemetryEvent(props: HoverOverlayProps): void {
    const { telemetryServiceV2, hoveredToken } = props

    const previousPropsReference = useRef(props)
    const recordTelemetryEvent = (): void => telemetryServiceV2.record(EventName.hover)

    // Record a telemetry event on component mount once, so we don't care about dependency updates.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => recordTelemetryEvent(), [])

    // Record a telemetry event for this hover being displayed,
    // but only do it once per position and when it is non-empty.
    if (
        !isEmptyHover(props) &&
        (!isEqual(hoveredToken, previousPropsReference.current.hoveredToken) ||
            isEmptyHover(previousPropsReference.current))
    ) {
        recordTelemetryEvent()
    }

    // Update previous props ref after we used it to determine if a telemetry event is needed.
    previousPropsReference.current = props
}
