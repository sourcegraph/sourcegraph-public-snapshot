import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

const notEnabled = {
    chat: false,
    sidebar: false,
    search: false,
    editorRecipes: false,
    needsEmailVerification: false,
}

interface IsCodyEnabled {
    chat: boolean
    sidebar: boolean
    search: boolean
    editorRecipes: boolean
    needsEmailVerification: boolean
}

export const isEmailVerificationNeeded = (): boolean =>
    window.context?.codyRequiresVerifiedEmail && !window.context?.currentUser?.hasVerifiedEmail

export const useIsCodyEnabled = (): IsCodyEnabled => {
    const [chatEnabled] = useFeatureFlag('cody-web-chat')
    const [searchEnabled] = useFeatureFlag('cody-web-search')
    const [sidebarEnabled] = useFeatureFlag('cody-web-sidebar')
    const [editorRecipesEnabled] = useFeatureFlag('cody-web-editor-recipes')
    let [allEnabled] = useFeatureFlag('cody-web-all')

    if (!window.context?.codyEnabled) {
        return notEnabled
    }
    if (window.context.sourcegraphAppMode) {
        // If the user is using the Sourcegraph app, all features are enabled
        // as long as the user has a connected Sourcegraph.com account.
        allEnabled = true
    }

    return {
        chat: chatEnabled || allEnabled,
        sidebar: sidebarEnabled || allEnabled,
        search: searchEnabled || allEnabled,
        editorRecipes: (editorRecipesEnabled && sidebarEnabled) || allEnabled,
        needsEmailVerification: isEmailVerificationNeeded(),
    }
}
