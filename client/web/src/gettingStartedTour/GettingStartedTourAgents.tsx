import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useCallback, useEffect, useState } from 'react'
import ReactDOM from 'react-dom'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useGettingStartedTourState } from '../stores/gettingStartedTourState'

import { GettingStartedTourStepItem } from './data'
import styles from './GettingStartedTour.module.scss'
import { GETTING_STARTED_TOUR_MARKER } from './GettingStartedTourInfo'
import { parseURIMarkers } from './utils'

interface GettingStartedTourCompletionAgentProps extends Partial<TelemetryProps> {
    steps: GettingStartedTourStepItem[]
}
/**
 * Agent component that tracks step completions
 */
export const GettingStartedTourCompletionAgent: React.FunctionComponent<GettingStartedTourCompletionAgentProps> = React.memo(
    ({ steps, telemetryService }) => {
        const addCompletedID = useGettingStartedTourState(useCallback(state => state.addCompletedID, []))

        useEffect(() => {
            const filteredSteps = steps.filter(step => step.completeAfterEvents)
            telemetryService?.addEventLogListener?.(eventName => {
                const stepId = filteredSteps.find(step => step.completeAfterEvents?.includes(eventName))?.id
                if (stepId) {
                    addCompletedID(stepId)
                }
            })
        }, [telemetryService, steps, addCompletedID])

        return null
    }
)
interface GettingStartedTourInfoAgentProps {
    steps: GettingStartedTourStepItem[]
}
/**
 * Agent component that shows info dialogs
 */
export const GettingStartedTourInfoAgent: React.FunctionComponent<GettingStartedTourInfoAgentProps> = React.memo(
    ({ steps }) => {
        const [info, setInfo] = useState<GettingStartedTourStepItem['info'] | undefined>()

        const location = useLocation()

        useEffect(() => {
            const { isTour, stepId } = parseURIMarkers(location.search)
            if (!isTour || !stepId) {
                return
            }

            const info = steps.find(step => stepId === step.id)?.info
            if (info) {
                setInfo(info)
            }
        }, [steps, location])

        if (!info) {
            return null
        }

        const domNode = document.querySelector('.' + GETTING_STARTED_TOUR_MARKER)
        if (!domNode) {
            return null
        }

        return ReactDOM.createPortal(
            <div className={styles.infoPanel}>
                <CheckCircleIcon className={classNames('icon-inline', styles.infoIcon)} size="1rem" />
                <span dangerouslySetInnerHTML={{ __html: info }} />
            </div>,
            domNode
        )
    }
)
