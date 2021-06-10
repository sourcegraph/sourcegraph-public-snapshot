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
    featureId: string,
    showFeatureTour: boolean,
    getFeatureTourElement: (onClose: () => void) => HTMLElement,
    localStorageKey: string,
    tourOptions: Shepherd.Tour.TourOptions
): Shepherd.Tour => {
    const [hasSeenFeatureTour, setHasSeenFeatureTour] = useLocalStorage(localStorageKey, false)

    const tour = useMemo(() => new Shepherd.Tour(tourOptions), [tourOptions])
    useEffect(() => {
        tour.addSteps([{ id: featureId, text: getFeatureTourElement(() => tour.cancel()) }])
    }, [tour, featureId, getFeatureTourElement])

    useEffect(() => {
        if (!tour.isActive() && showFeatureTour && !hasSeenFeatureTour) {
            tour.start()
        }
    }, [showFeatureTour, hasSeenFeatureTour, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenFeatureTour(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenFeatureTour])

    useEffect(() => () => tour.cancel(), [tour])

    return tour
}
