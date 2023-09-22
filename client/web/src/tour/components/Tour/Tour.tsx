import React, { useCallback, useEffect, useMemo, useState } from 'react'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TourContext } from './context'
import { TourAgent } from './TourAgent'
import { TourContent } from './TourContent'
import { useTour } from './useTour'
import { isQuerySuccessful } from './utils'

export type TourProps = TelemetryProps & {
    id: string
    tasks: TourTaskType[]
    extraTask?: TourTaskType
    defaultSnippets: Record<string, string[]>
} & Pick<React.ComponentProps<typeof TourContent>, 'variant' | 'className' | 'height' | 'title' | 'keepCompletedTasks'>

export const Tour: React.FunctionComponent<React.PropsWithChildren<TourProps>> = React.memo(
    ({ id: tourId, tasks, extraTask, defaultSnippets, telemetryService, ...props }) => {
        const { completedStepIds = [], status, setStepCompleted, setStatus, restart } = useTour(tourId)
        const onLogEvent = useCallback(
            (eventName: string, eventProperties?: any, publicArgument?: any) => {
                telemetryService.log(tourId + eventName, { ...eventProperties }, { ...publicArgument })
            },
            [telemetryService, tourId]
        )

        // TODO: Read these values from user config
        const [userorg] = useState<string>('sourcegraph')
        const [userrepo] = useState<string>('sourcegraph')
        const [userlang] = useState<string>('go')

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
                    const extendedSteps = task.steps.map(step => {
                        const extendedStep = {
                            ...step,
                            isCompleted: completedStepIds.includes(step.id),
                        }

                        switch (extendedStep.action.type) {
                            case 'search-query':
                                if (!extendedStep.action.snippets) {
                                    extendedStep.action = {
                                        ...extendedStep.action,
                                        snippets: defaultSnippets,
                                    }
                                }
                                break
                        }

                        return extendedStep
                    })

                    return {
                        ...task,
                        steps: extendedSteps,
                        completed: Math.round(
                            (100 * extendedSteps.filter(step => step.isCompleted).length) /
                                (task.requiredSteps ?? extendedSteps.length)
                        ),
                    }
                }),
            [tasks, completedStepIds, defaultSnippets]
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
            <TourContext.Provider
                value={{ onStepClick, onRestart, userConfig: { userlang, userorg, userrepo }, isQuerySuccessful }}
            >
                <TourContent {...props} onClose={onClose} tasks={extendedTasks} />
                <TourAgent tasks={finalTasks} telemetryService={telemetryService} onStepComplete={onStepComplete} />
            </TourContext.Provider>
        )
    }
)

Tour.displayName = 'Tour'
