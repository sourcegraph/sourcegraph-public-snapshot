import React, { useCallback, useEffect, useMemo } from 'react'

import type { TourTaskStepType, TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import type { UserOnboardingConfig } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TourContext } from './context'
import { TourAgent } from './TourAgent'
import { TourContent } from './TourContent'
import { useTour } from './useTour'
import { canRunStep, isNotNullOrUndefined, isQuerySuccessful } from './utils'

// Ensure tour names are known strings
type tourIds = 'MockTour' | 'GettingStarted' | 'TourStorybook'

export type TourProps = TelemetryProps &
    TelemetryV2Props & {
        id: tourIds
        tasks: TourTaskType[]
        extraTask?: TourTaskType
        userInfo?: UserOnboardingConfig['userinfo']
        defaultSnippets: Record<string, string[]>
    } & Pick<
        React.ComponentProps<typeof TourContent>,
        'variant' | 'className' | 'height' | 'title' | 'keepCompletedTasks'
    >

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
            telemetryRecorder.recordEvent(`tour.${tourId}`, 'view')
        }, [onLogEvent, telemetryRecorder, tourId])

        const onClose = useCallback(() => {
            onLogEvent('Closed')
            telemetryRecorder.recordEvent(`tour.${tourId}`, 'close')
            setStatus('closed')
        }, [onLogEvent, telemetryRecorder, tourId, setStatus])

        const onStepComplete = useCallback(
            (step: TourTaskStepType) => {
                setStepCompleted(step.id)
            },
            [setStepCompleted]
        )

        const onStepClick = useCallback(
            (step: TourTaskStepType) => {
                onLogEvent('StepClicked', { stepId: step.id })
                telemetryRecorder.recordEvent(`tour.${tourId}.step`, 'click')
                if (step.completeAfterEvents) {
                    return
                }
                onStepComplete(step)
            },
            [onLogEvent, telemetryRecorder, tourId, onStepComplete]
        )

        const onRestart = useCallback(
            (step: TourTaskStepType) => {
                onLogEvent('RestartClicked')
                telemetryRecorder.recordEvent(`tour.${tourId}`, 'restart')
                restart()
            },
            [onLogEvent, telemetryRecorder, tourId, restart]
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
                                    case 'search-query': {
                                        if (!extendedStep.action.snippets) {
                                            extendedStep.action = {
                                                ...extendedStep.action,
                                                snippets: defaultSnippets,
                                            }
                                        }
                                        break
                                    }
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
                telemetryRecorder.recordEvent(`tour.${tourId}`, 'complete')
                setStatus('completed')
            }
        }, [status, extendedTasks, onLogEvent, telemetryRecorder, setStatus, tourId])

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
                <TourAgent tasks={finalTasks} telemetryService={telemetryService} onStepComplete={onStepComplete} />
            </TourContext.Provider>
        )
    }
)

Tour.displayName = 'Tour'
