import React, { useEffect, useRef, useState, useMemo } from 'react'

import { mdiCheckCircle } from '@mdi/js'
import ReactDOM from 'react-dom'
import { useLocation } from 'react-router-dom'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon } from '@sourcegraph/wildcard'

import { GETTING_STARTED_TOUR_MARKER } from './TourInfo'
import { parseURIMarkers } from './utils'

import styles from './Tour.module.scss'

interface TourAgentProps extends TelemetryProps, TelemetryV2Props {
    tasks: TourTaskType[]
    onStepComplete: (step: TourTaskStepType) => void
}

export function useTourQueryParameters(): ReturnType<typeof parseURIMarkers> {
    const location = useLocation()
    return useMemo(() => parseURIMarkers(location.search), [location])
}

/**
 * Component to track TourTaskStepType.completeAfterEvents and show info box for steps.
 */
export const TourAgent: React.FunctionComponent<React.PropsWithChildren<TourAgentProps>> = React.memo(
    ({ tasks, telemetryService, onStepComplete }) => {
        // Agent 1: Track completion
        useEffect(() => {
            const filteredSteps = tasks.flatMap(task => task.steps).filter(step => step.completeAfterEvents)
            return telemetryService?.addEventLogListener?.(eventName => {
                const step = filteredSteps.find(step => step.completeAfterEvents?.includes(eventName))
                if (step) {
                    onStepComplete(step)
                }
            })
        }, [telemetryService, tasks, onStepComplete])

        // Agent 2: Track info panel
        const [info, setInfo] = useState<TourTaskStepType['info'] | undefined>()

        const tourQueryParameters = useTourQueryParameters()

        useEffect(() => {
            const info = tasks.flatMap(task => task.steps).find(step => tourQueryParameters.stepId === step.id)?.info
            setInfo(info)
        }, [tasks, tourQueryParameters.stepId])

        const infoContainerReference = useRef(document.querySelector('#' + GETTING_STARTED_TOUR_MARKER))
        useEffect(() => {
            if (info) {
                infoContainerReference.current?.classList.remove('d-none')
            } else {
                infoContainerReference.current?.classList.add('d-none')
            }
        }, [info])

        if (!info || !infoContainerReference.current) {
            return null
        }

        return ReactDOM.createPortal(
            <div className={styles.infoPanel}>
                <Icon className={styles.infoIcon} aria-hidden={true} svgPath={mdiCheckCircle} />
                <span>{info}</span>
            </div>,
            infoContainerReference.current
        )
    }
)
