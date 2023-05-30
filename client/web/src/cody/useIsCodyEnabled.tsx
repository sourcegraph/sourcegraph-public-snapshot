import { useMemo } from 'react'

import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

export const notEnabled = {
    loaded: true,
    chat: false,
    sidebar: false,
    search: false,
    editorRecipes: false,
    needsEmailVerification: false,
}

export interface IsCodyEnabled {
    loaded: boolean
    chat: boolean
    sidebar: boolean
    search: boolean
    editorRecipes: boolean
    needsEmailVerification: boolean
}

export const isEmailVerificationNeeded = (): boolean =>
    window.context?.codyRequiresVerifiedEmail && !window.context?.currentUser?.hasVerifiedEmail

export const useIsCodyEnabled = (): IsCodyEnabled => {
    const [chatEnabled, chatEnabledStatus] = useFeatureFlag('cody-web-chat')
    const [searchEnabled, searchEnabledStatus] = useFeatureFlag('cody-web-search')
    const [sidebarEnabled, sidebarEnabledStatus] = useFeatureFlag('cody-web-sidebar')
    const [editorRecipesEnabled, editorRecipesEnabledStatus] = useFeatureFlag('cody-web-editor-recipes')
    let [allEnabled, allEnabledStatus] = useFeatureFlag('cody-web-all')

    if (window.context?.sourcegraphAppMode) {
        // If the user is using the Sourcegraph app, all features are enabled
        // as long as the user has a connected Sourcegraph.com account.
        allEnabled = true
    }

    const enabled = useMemo(
        () => ({
            loaded:
                window.context?.sourcegraphAppMode ||
                (chatEnabledStatus === 'loaded' &&
                    searchEnabledStatus === 'loaded' &&
                    sidebarEnabledStatus === 'loaded' &&
                    editorRecipesEnabledStatus === 'loaded' &&
                    allEnabledStatus === 'loaded'),
            chat: chatEnabled || allEnabled,
            sidebar: sidebarEnabled || allEnabled,
            search: searchEnabled || allEnabled,
            editorRecipes: (editorRecipesEnabled && sidebarEnabled) || allEnabled,
            needsEmailVerification: isEmailVerificationNeeded(),
        }),
        [
            chatEnabled,
            sidebarEnabled,
            searchEnabled,
            editorRecipesEnabled,
            allEnabled,
            chatEnabledStatus,
            searchEnabledStatus,
            sidebarEnabledStatus,
            editorRecipesEnabledStatus,
            allEnabledStatus,
        ]
    )

    if (!window.context?.codyEnabled) {
        return notEnabled
    }

    return enabled
}
