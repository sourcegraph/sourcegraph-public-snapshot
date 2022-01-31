import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useCallback, useEffect, useState } from 'react'
import ReactDOM from 'react-dom'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useOnboardingTourState } from '../stores/onboardingTourState'

import { OnboardingTourStepItem } from './data'
import styles from './OnboardingTour.module.scss'
import { ONBOARDING_TOUR_MARKER } from './OnboardingTourInfo'
import { parseURIMarkers } from './utils'

interface OnboardingTourCompletionAgentProps extends Partial<TelemetryProps> {
    steps: OnboardingTourStepItem[]
}
/**
 * Agent component that tracks step completions
 */
export const OnboardingTourCompletionAgent: React.FunctionComponent<OnboardingTourCompletionAgentProps> = React.memo(
    ({ steps, telemetryService }) => {
        const addCompletedID = useOnboardingTourState(useCallback(state => state.addCompletedID, []))

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
interface OnboardingTourInfoAgentProps {
    steps: OnboardingTourStepItem[]
}
/**
 * Agent component that shows info dialogs
 */
export const OnboardingTourInfoAgent: React.FunctionComponent<OnboardingTourInfoAgentProps> = React.memo(
    ({ steps }) => {
        const [info, setInfo] = useState<OnboardingTourStepItem['info'] | undefined>()

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

        const domNode = document.querySelector('.' + ONBOARDING_TOUR_MARKER)
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
