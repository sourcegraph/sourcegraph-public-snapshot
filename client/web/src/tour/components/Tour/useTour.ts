import { useCallback } from 'react'

import { omit, uniq } from 'lodash'
import type { Optional } from 'utility-types'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

export interface TourState {
    completedStepIds?: string[]
    status?: 'closed' | 'completed'
}

export type TourListState = Optional<Record<string, TourState>>

export type UseTourReturnType = TourState & {
    setStepCompleted: (stepId: string) => void
    setStatus: (status: TourState['status']) => void
    restart: () => void
}

export function useTour(tourKey: string): UseTourReturnType {
    const [allToursSate, setAllToursState, loadStatus] = useTemporarySetting('onboarding.quickStartTour')

    const setStepCompleted = useCallback(
        (stepId: string): void =>
            setAllToursState(previousState => ({
                ...previousState,
                [tourKey]: {
                    ...previousState?.[tourKey],
                    completedStepIds: uniq([...(previousState?.[tourKey]?.completedStepIds ?? []), stepId]),
                },
            })),
        [setAllToursState, tourKey]
    )

    const setStatus = useCallback(
        (status: TourState['status']): void =>
            setAllToursState(previousState => ({
                ...previousState,
                [tourKey]: { ...previousState?.[tourKey], status },
            })),
        [setAllToursState, tourKey]
    )

    const restart = useCallback(
        (): void => setAllToursState(previousState => omit(previousState, tourKey)),
        [setAllToursState, tourKey]
    )

    return {
        ...allToursSate?.[tourKey],
        // To avoid rendering Tour.tsx when state is still loading
        ...(loadStatus !== 'loaded' ? { status: 'closed' } : {}),
        setStepCompleted,
        setStatus,
        restart,
    }
}
