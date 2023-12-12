import React, { useCallback, useEffect, useMemo } from 'react'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import type { UserOnboardingConfig } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TourContext } from './context'
import { TourAgent } from './TourAgent'
import { TourContent } from './TourContent'
import { useTour } from './useTour'
import { canRunStep, isNotNullOrUndefined, isQuerySuccessful } from './utils'

export type TourProps = TelemetryProps & {
    id: string
    tasks: TourTaskType[]
    extraTask?: TourTaskType
    userInfo?: UserOnboardingConfig['userinfo']
    defaultSnippets: Record<string, string[]>
} & Pick<React.ComponentProps<typeof TourContent>, 'variant' | 'className' | 'height' | 'title' | 'keepCompletedTasks'>

export const Tour: React.FunctionComponent<React.PropsWithChildren<TourProps>> = React.memo(
    ({ id: tourId, tasks, extraTask, defaultSnippets, telemetryService, telemetryRecorder, userInfo, ...props }) => {
        const { completedStepIds = [], status, setStepCompleted, setStatus, restart } = useTour(tourId)
        const onLogEvent = useCallback(
            (eventName: string, eventProperties?: any, publicArgument?: any) => {
                telemetryService.log('Tour' + eventName, { tourId, ...eventProperties }, { ...publicArgument })
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
                onLogEvent('StepClicked', { stepId: step.id })
                if (step.completeAfterEvents) {
                    return
                }
                onStepComplete(step)
            },
            [onLogEvent, onStepComplete]
        )

        const onRestart = useCallback(
            (step: TourTaskStepType) => {
                onLogEvent('RestartClicked')
                restart()
            },
            [onLogEvent, restart]
        )

        const extendedTasks: TourTaskType[] = useMemo(
            () =>
                tasks
                    .map(task => {
                        const extendedSteps = task.steps
                            .filter(step => canRunStep(step, userInfo))
                            .map(step => {
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

                        if (extendedSteps.length === 0) {
                            return null
                        }

                        return {
                            ...task,
                            steps: extendedSteps,
                            completed: Math.round(
                                (100 * extendedSteps.filter(step => step.isCompleted).length) /
                                    (task.requiredSteps ?? extendedSteps.length)
                            ),
                        }
                    })
                    .filter(isNotNullOrUndefined),
            [tasks, completedStepIds, defaultSnippets, userInfo]
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
            <TourContext.Provider value={{ onStepClick, onRestart, userInfo, isQuerySuccessful }}>
                <TourContent {...props} onClose={onClose} tasks={finalTasks} />
                <TourAgent
                    tasks={finalTasks}
                    telemetryService={telemetryService}
                    telemetryRecorder={telemetryRecorder}
                    onStepComplete={onStepComplete}
                />
            </TourContext.Provider>
        )
    }
)

Tour.displayName = 'Tour'
