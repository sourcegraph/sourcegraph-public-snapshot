import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

export const useArePaymentsEnabled = (): boolean => {
    const [enabled] = useFeatureFlag('use-ssc-for-cody-subscription', false)

    return enabled
}

export const useHasTrialEnded = (): boolean => {
    const [ended] = useFeatureFlag('cody-pro-trial-ended', false)

    return ended
}
