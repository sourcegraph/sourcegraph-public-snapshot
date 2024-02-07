import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

export const useArePaymentsEnabled = (): boolean => {
    const [enabled] = useFeatureFlag('use-ssc-for-cody-subscription-on-web', false)

    return enabled
}

export const useHasTrialEnded = (): boolean => {
    const [ended] = useFeatureFlag('cody-pro-trial-ended', false)

    return ended
}

// TODO(sourcegraph:#60213) remove this flag after testing is complete
export const useIsCodyPaymentsTestingMode = (): boolean => {
    const [testingMode] = useFeatureFlag('cody-payments-testing-mode', false)

    return testingMode
}
