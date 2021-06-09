import { useEffect, useMemo } from 'react'
import Shepherd from 'shepherd.js'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import styles from './FeatureTour.module.scss'
import { defaultTourOptions } from './input/tour-options'

export const getTourOptions = (stepOptions: Shepherd.Step.StepOptions): Shepherd.Tour.TourOptions => ({
    ...defaultTourOptions,
    defaultStepOptions: {
        ...defaultTourOptions.defaultStepOptions,
        arrow: true,
        classes: `web-content shadow-lg py-4 px-3 ${styles.featureTour}`,
        ...stepOptions,
    },
})

export const useFeatureTour = (
    featureStepId: string,
    showFeatureTour: boolean,
    getFeatureTourStepElement: (onClose: () => void) => HTMLElement,
    localStorageKey: string,
    tourOptions: Shepherd.Tour.TourOptions
): Shepherd.Tour => {
    const [hasSeenFeatureTourStep, setHasSeenFeatureTourStep] = useLocalStorage(localStorageKey, false)

    const tour = useMemo(() => new Shepherd.Tour(tourOptions), [tourOptions])
    useEffect(() => {
        tour.addSteps([{ id: featureStepId, text: getFeatureTourStepElement(() => tour.cancel()) }])
    }, [tour, featureStepId, getFeatureTourStepElement])

    useEffect(() => {
        if (!tour.isActive() && showFeatureTour && !hasSeenFeatureTourStep) {
            tour.start()
        }
    }, [showFeatureTour, hasSeenFeatureTourStep, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenFeatureTourStep(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenFeatureTourStep])

    useEffect(() => () => tour.cancel(), [tour])

    return tour
}
