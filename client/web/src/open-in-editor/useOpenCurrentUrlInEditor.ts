import { useCallback } from 'react'

import { useLocation } from 'react-router'

import { parseBrowserRepoURL } from '../util/url'

import { buildEditorUrl, buildRepoBaseNameAndPath } from './build-url'
import { EditorSettings } from './editor-settings'

/**
 * @returns A function to open the current location in your preferred editor
 */
export const useOpenCurrentUrlInEditor = (): ((
    editorSettings: EditorSettings | undefined,
    externalServiceType: string | undefined,
    sourcegraphURL: string,
    editorIndex?: number
) => void) => {
    const location = useLocation()
    return useCallback(
        (
            editorSettings: EditorSettings | undefined,
            externalServiceType: string | undefined,
            sourcegraphURL: string,
            editorIndex = 0
        ) => {
            const { repoName, filePath, range } = parseBrowserRepoURL(location.pathname)
            const url = buildEditorUrl(
                buildRepoBaseNameAndPath(repoName, externalServiceType, filePath),
                range,
                editorSettings,
                sourcegraphURL,
                editorIndex
            )
            window.open(url.toString(), '_blank')
        },
        [location.pathname]
    )
}
