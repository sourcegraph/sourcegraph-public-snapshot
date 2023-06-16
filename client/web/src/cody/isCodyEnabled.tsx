export const isEmailVerificationNeededForCody = (): boolean =>
    window.context?.codyRequiresVerifiedEmail && !window.context?.currentUser?.hasVerifiedEmail

export const isCodyEnabled = (): boolean => {
    if (window.context?.sourcegraphAppMode) {
        return true
    }

    if (!window.context?.codyEnabled || !window.context?.codyEnabledForCurrentUser) {
        return false
    }

    return true
}
