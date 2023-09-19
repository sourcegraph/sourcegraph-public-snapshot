import React, { useCallback, useEffect, useMemo } from 'react'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TourContext } from './context'
import { TourAgent } from './TourAgent'
import { TourContent } from './TourContent'
import { useTour } from './useTour'

export type TourProps = TelemetryProps & {
    id: string
    tasks: TourTaskType[]
    extraTask?: TourTaskType
} & Pick<React.ComponentProps<typeof TourContent>, 'variant' | 'className' | 'height' | 'title' | 'keepCompletedTasks'>

export const Tour: React.FunctionComponent<React.PropsWithChildren<TourProps>> = React.memo(
    ({ id: tourId, tasks, extraTask, telemetryService, ...props }) => {
        const { completedStepIds = [], status, setStepCompleted, setStatus, restart } = useTour(tourId)
        const onLogEvent = useCallback(
            (eventName: string, eventProperties?: any, publicArgument?: any) => {
                telemetryService.log(tourId + eventName, { ...eventProperties }, { ...publicArgument })
            },
            [telemetryService, tourId]
        )

        useEffect(() => {
            onLogEvent('Shown')
        }, [onLogEvent, tourId])

        const onClose = useCallback(() => {
            onLogEvent('Closed')
            setStatus('closed')
        }, [onLogEvent, setStatus])

        const onStepComplete = useCallback(
            (step: TourTaskStepType) => {
                setStepCompleted(step.id)
            },
            [setStepCompleted]
        )

        const onStepClick = useCallback(
            (step: TourTaskStepType) => {
                onLogEvent(step.id + 'Clicked')
                if (step.completeAfterEvents) {
                    return
                }
                onStepComplete(step)
            },
            [onLogEvent, onStepComplete]
        )

        const onRestart = useCallback(
            (step: TourTaskStepType) => {
                onLogEvent(step.id + 'Clicked')
                restart()
            },
            [onLogEvent, restart]
        )

        const extendedTasks: TourTaskType[] = useMemo(
            () =>
                tasks.map(task => {
                    const extendedSteps = task.steps.map(step => ({
                        ...step,
                        isCompleted: completedStepIds.includes(step.id),
                    }))

                    return {
                        ...task,
                        steps: extendedSteps,
                        completed: Math.round(
                            (100 * extendedSteps.filter(step => step.isCompleted).length) /
                                (task.requiredSteps ?? extendedSteps.length)
                        ),
                    }
                }),
            [tasks, completedStepIds]
        )

        useEffect(() => {
            if (
                status !== 'closed' &&
                status !== 'completed' &&
                extendedTasks.filter(task => task.completed === 100).length === extendedTasks.length
            ) {
                onLogEvent('Completed')
                setStatus('completed')
            }
        }, [status, extendedTasks, onLogEvent, setStatus, tourId])

        if (status === 'closed') {
            return null
        }

        const finalTasks = [...extendedTasks]
        if (status === 'completed' && extraTask) {
            finalTasks.unshift(extraTask)
        }

        return (
            <TourContext.Provider value={{ onStepClick, onRestart }}>
                <TourContent {...props} onClose={onClose} tasks={extendedTasks} />
                <TourAgent tasks={finalTasks} telemetryService={telemetryService} onStepComplete={onStepComplete} />
            </TourContext.Provider>
        )
    }
)

Tour.displayName = 'Tour'
