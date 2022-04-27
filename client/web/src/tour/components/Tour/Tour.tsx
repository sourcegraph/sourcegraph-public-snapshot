import React, { useCallback, useEffect, useMemo } from 'react'

import { uniq } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useQuickStartTourListState } from '../../../stores/quickStartTourState'

import { TourContext } from './context'
import { TourAgent } from './TourAgent'
import { TourContent } from './TourContent'
import { TourTaskType, TourLanguage, TourTaskStepType } from './types'
import { isLanguageRequired } from './utils'

export type TourProps = TelemetryProps & {
    id: string
    tasks: TourTaskType[]
    extraTask?: TourTaskType
} & Pick<React.ComponentProps<typeof TourContent>, 'variant' | 'className' | 'height'>

export const Tour: React.FunctionComponent<TourProps> = ({
    id: tourId,
    tasks,
    extraTask,
    telemetryService,
    ...props
}) => {
    const {
        completedStepIds = [],
        language,
        status,
        setLanguage,
        setCompletedStepIds,
        setStatus,
        resetTour,
    } = useQuickStartTourListState(
        useCallback(
            ({ tours, setLanguage, setCompletedStepIds, setStatus, resetTour }) => ({
                ...tours[tourId],
                setLanguage,
                setCompletedStepIds,
                setStatus,
                resetTour,
            }),
            [tourId]
        )
    )
    const onLogEvent = useCallback(
        (eventName: string, eventProperties?: any, publicArgument?: any) => {
            telemetryService.log(tourId + eventName, { language, ...eventProperties }, { language, ...publicArgument })
        },
        [language, telemetryService, tourId]
    )

    useEffect(() => {
        onLogEvent('Shown')
    }, [onLogEvent, tourId])

    const onClose = useCallback(() => {
        onLogEvent('Closed')
        setStatus(tourId, 'closed')
    }, [onLogEvent, setStatus, tourId])

    const onStepComplete = useCallback(
        (step: TourTaskStepType) => {
            const newCompletedStepIds = uniq([...completedStepIds, step.id])
            setCompletedStepIds(tourId, newCompletedStepIds)
        },
        [completedStepIds, setCompletedStepIds, tourId]
    )

    const onStepClick = useCallback(
        (step: TourTaskStepType, language?: TourLanguage) => {
            onLogEvent(step.id + 'Clicked', { language }, { language })
            if (step.completeAfterEvents || (isLanguageRequired(step) && !language)) {
                return
            }
            onStepComplete(step)
        },
        [onLogEvent, onStepComplete]
    )

    const onLanguageSelect = useCallback(
        (language: TourLanguage) => {
            setLanguage(tourId, language)
            onLogEvent('LanguageClicked', { language }, { language })
        },
        [onLogEvent, setLanguage, tourId]
    )

    const onRestart = useCallback(
        (step: TourTaskStepType) => {
            onLogEvent(step.id + 'Clicked')
            resetTour(tourId)
        },
        [onLogEvent, resetTour, tourId]
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
                        (100 * extendedSteps.filter(step => step.isCompleted).length) / extendedSteps.length
                    ),
                }
            }),
        [tasks, completedStepIds]
    )

    useEffect(() => {
        if (
            !['completed', 'closed'].includes(status as string) &&
            extendedTasks.filter(step => step.completed === 100).length === extendedTasks.length
        ) {
            onLogEvent('Completed')
            setStatus(tourId, 'completed')
        }
    }, [status, extendedTasks, onLogEvent, setStatus, tourId])

    if (status === 'closed') {
        return null
    }

    return (
        <TourContext.Provider value={{ onStepClick, language, onLanguageSelect, onRestart }}>
            <TourContent
                {...props}
                onClose={onClose}
                tasks={
                    [...extendedTasks, status === 'completed' && extraTask].filter(Boolean) as (
                        | TourTaskType
                        | TourTaskType
                    )[]
                }
            />
            <TourAgent tasks={extendedTasks} telemetryService={telemetryService} onStepComplete={onStepComplete} />
        </TourContext.Provider>
    )
}
