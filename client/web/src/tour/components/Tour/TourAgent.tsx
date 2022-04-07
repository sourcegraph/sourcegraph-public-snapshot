import React, { useEffect, useState } from 'react'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ReactDOM from 'react-dom'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon } from '@sourcegraph/wildcard'

import { GETTING_STARTED_TOUR_MARKER } from './TourInfo'
import { TourTaskType, TourTaskStepType } from './types'
import { parseURIMarkers } from './utils'

import styles from './Tour.module.scss'

interface TourAgentProps extends TelemetryProps {
    tasks: TourTaskType[]
    onStepComplete: (step: TourTaskStepType) => void
}

/**
 * Component to track TourTaskStepType.completeAfterEvents and show info box for steps.
 */
export const TourAgent: React.FunctionComponent<TourAgentProps> = React.memo(
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

        const location = useLocation()

        useEffect(() => {
            const { isTour, stepId } = parseURIMarkers(location.search)
            if (!isTour || !stepId) {
                return
            }

            const info = tasks.flatMap(task => task.steps).find(step => stepId === step.id)?.info
            if (info) {
                setInfo(info)
            }
        }, [tasks, location])

        if (!info) {
            return null
        }

        const domNode = document.querySelector('.' + GETTING_STARTED_TOUR_MARKER)
        if (!domNode) {
            return null
        }

        return ReactDOM.createPortal(
            <div className={styles.infoPanel}>
                <Icon as={CheckCircleIcon} className={styles.infoIcon} />
                <span>{info}</span>
            </div>,
            domNode
        )
    }
)
