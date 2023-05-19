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
    window.context?.sourcegraphDotComMode && !window.context?.currentUser?.hasVerifiedEmail

export const useIsCodyEnabled = (): IsCodyEnabled => {
    const [chatEnabled] = useFeatureFlag('cody-web-chat')
    const [searchEnabled] = useFeatureFlag('cody-web-search')
    const [sidebarEnabled] = useFeatureFlag('cody-web-sidebar')
    const [editorRecipesEnabled] = useFeatureFlag('cody-web-editor-recipes')
    const [allEnabled] = useFeatureFlag('cody-web-all')

    if (!window.context?.codyEnabled) {
        return notEnabled
    }

    return {
        chat: chatEnabled || allEnabled,
        sidebar: sidebarEnabled || allEnabled,
        search: searchEnabled || allEnabled,
        editorRecipes: (editorRecipesEnabled && sidebarEnabled) || allEnabled,
        needsEmailVerification: isEmailVerificationNeeded(),
    }
}
