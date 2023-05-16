import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

const notEnabled = {
    chat: false,
    sidebar: false,
    search: false,
    editorRecipes: false,
}

export const useIsCodyEnabled = (): { chat: boolean; sidebar: boolean; search: boolean; editorRecipes: boolean } => {
    const [chatEnabled] = useFeatureFlag('cody-web-chat')
    const [searchEnabled] = useFeatureFlag('cody-web-search')
    const [sidebarEnabled] = useFeatureFlag('cody-web-sidebar')
    const [editorRecipesEnabled] = useFeatureFlag('cody-web-editor-recipes')
    const [allEnabled] = useFeatureFlag('cody-web-all')

    if (!window.context?.codyEnabled) {
        return notEnabled
    }

    if (
        window.context?.sourcegraphDotComMode &&
        !window.context?.currentUser?.siteAdmin &&
        !window.context?.currentUser?.hasVerifiedEmail
    ) {
        return notEnabled
    }

    return {
        chat: chatEnabled || allEnabled,
        sidebar: sidebarEnabled || allEnabled,
        search: searchEnabled || allEnabled,
        editorRecipes: (editorRecipesEnabled && sidebarEnabled) || allEnabled,
    }
}
