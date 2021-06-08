import { useEffect, useMemo } from 'react'
import Shepherd from 'shepherd.js'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { defaultTourOptions } from './input/tour-options'

export const getTourOptions = (stepOptions: Shepherd.Step.StepOptions): Shepherd.Tour.TourOptions => ({
    ...defaultTourOptions,
    defaultStepOptions: {
        ...defaultTourOptions.defaultStepOptions,
        arrow: true,
        classes: 'web-content shadow-lg py-4 px-3 highlight-tour',
        ...stepOptions,
    },
})

export const useHighlightTour = (
    highlightStepId: string,
    showHighlightStep: boolean,
    getHighlightStepElement: (onClose: () => void) => HTMLElement,
    localStorageKey: string,
    tourOptions: Shepherd.Tour.TourOptions
): Shepherd.Tour => {
    const [hasSeenHighlightTourStep, setHasSeenHighlightTourStep] = useLocalStorage(localStorageKey, false)

    const tour = useMemo(() => new Shepherd.Tour(tourOptions), [tourOptions])
    useEffect(() => {
        tour.addSteps([{ id: highlightStepId, text: getHighlightStepElement(() => tour.cancel()) }])
    }, [tour, highlightStepId, getHighlightStepElement])

    useEffect(() => {
        if (!tour.isActive() && showHighlightStep && !hasSeenHighlightTourStep) {
            tour.start()
        }
    }, [showHighlightStep, hasSeenHighlightTourStep, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenHighlightTourStep(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenHighlightTourStep])

    useEffect(() => () => tour.cancel(), [tour])

    return tour
}
