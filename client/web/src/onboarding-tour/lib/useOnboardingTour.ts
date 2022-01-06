import { useCallback, useEffect, useMemo } from 'react'
import { BehaviorSubject } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { browserExtensionInstalled } from '../../tracking/analyticsUtils'

import { OnboardingTourStepItem, ONBOARDING_STEP_ITEMS } from './data'

const STORAGE_KEY = 'ONBOARDING_TOUR'

interface OnboardingTourStore {
    completedStepIds?: string[]
    isClosed?: boolean
    isCompleteEventTriggered?: boolean
}

const data = new BehaviorSubject<OnboardingTourStore>(
    JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}') as OnboardingTourStore
)

// eslint-disable-next-line rxjs/no-ignored-subscription
data.subscribe(data => localStorage.setItem(STORAGE_KEY, JSON.stringify(data)))

const clear = (): void => {
    localStorage.removeItem(STORAGE_KEY)
    data.next({})
}

const set = <K extends keyof OnboardingTourStore>(key: K, value: OnboardingTourStore[K]): void => {
    data.next({ ...data.value, [key]: value })
}

export function useOnboardingTourTracking(
    telemetryService: TelemetryProps['telemetryService']
): {
    triggerEvent: (eventName: string) => void
} {
    const triggerEvent = useCallback(
        (eventName: string) => {
            const args = { language: 'Go', page: location.href }
            telemetryService.log(eventName, args, args)
        },
        [telemetryService]
    )

    return { triggerEvent }
}

export function useOnboardingTourCompletion(): {
    onStepComplete: (step: OnboardingTourStepItem) => void
} {
    const store = useObservable(data)

    const onStepComplete = useCallback(
        (step: OnboardingTourStepItem) => {
            set('completedStepIds', [...(store?.completedStepIds ?? []), step.id])
        },
        [store?.completedStepIds]
    )

    return { onStepComplete }
}

export function useOnboardingTour(
    telemetryService: TelemetryProps['telemetryService']
): {
    steps: OnboardingTourStepItem[]
    isTourCompleted: boolean
    isClosed: boolean
    onClose: () => void
    onRestart: () => void
    onSignUp: () => void
    onInstall: () => void
} {
    const store = useObservable(data)
    const { triggerEvent } = useOnboardingTourTracking(telemetryService)

    const onClose = useCallback(() => {
        set('isClosed', true)
    }, [])

    const onSignUp = useCallback(() => {
        triggerEvent('TourSignUpClicked')
    }, [triggerEvent])

    const onInstall = useCallback(() => {
        triggerEvent('TourInstallLocallyClicked')
    }, [triggerEvent])

    const onRestart = useCallback(() => {
        clear()
    }, [])

    const isBrowserExtensionInstalled = useObservable(browserExtensionInstalled)

    const { steps, isTourCompleted, isClosed } = useMemo(() => {
        const steps = ONBOARDING_STEP_ITEMS.map(step => ({
            ...step,
            isCompleted:
                step.id === 'TourBrowserExtensions'
                    ? !!isBrowserExtensionInstalled
                    : !!store?.completedStepIds?.includes(step.id),
        }))
        return {
            steps,
            isTourCompleted: steps.filter(step => step.isCompleted).length === steps.length,
            isClosed: !!store?.isClosed,
        }
    }, [store, isBrowserExtensionInstalled])

    useEffect(() => {
        if (!store) {
            return
        }
        if (isTourCompleted && !store.isCompleteEventTriggered) {
            triggerEvent('TourComplete')
            set('isCompleteEventTriggered', true)
        }
    }, [isTourCompleted, store, triggerEvent])

    return { steps, isTourCompleted, isClosed, onClose, onRestart, onInstall, onSignUp }
}
