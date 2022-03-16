import { useEffect, useMemo } from 'react'

import Shepherd from 'shepherd.js'

import { useLocalStorage } from '@sourcegraph/wildcard'

import { defaultTourOptions } from './input/tour-options'

import styles from './FeatureTour.module.scss'

export const HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY = 'has-seen-search-contexts-dropdown-highlight-tour-step'
export const HAS_SEEN_CODE_MONITOR_FEATURE_TOUR_KEY = 'has-seen-create-code-monitor-feature-tour'

export const getTourOptions = (stepOptions: Shepherd.Step.StepOptions): Shepherd.Tour.TourOptions => ({
    ...defaultTourOptions,
    defaultStepOptions: {
        ...defaultTourOptions.defaultStepOptions,
        arrow: true,
        classes: `shadow-lg py-4 px-3 ${styles.featureTour}`,
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

    // tourOptions are excluded from deps array because we need to keep the reference to the same tour instance
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const tour = useMemo(() => new Shepherd.Tour(tourOptions), [])
    useEffect(() => {
        tour.addSteps([{ id: featureId, text: getFeatureTourElement(() => tour.cancel()) }])
    }, [tour, featureId, getFeatureTourElement])

    useEffect(() => {
        if (!tour.isActive() && showFeatureTour && !hasSeenFeatureTour) {
            tour.start()
        }
    }, [featureId, showFeatureTour, hasSeenFeatureTour, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenFeatureTour(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenFeatureTour])

    useEffect(
        () => () => {
            tour.hide()
            if (tour.isActive()) {
                tour.cancel()
            }
        },
        [tour]
    )

    return tour
}
