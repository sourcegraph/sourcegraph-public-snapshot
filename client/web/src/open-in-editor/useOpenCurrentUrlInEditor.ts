import { useCallback } from 'react'

import { useLocation } from 'react-router'

import { parseBrowserRepoURL } from '../util/url'

import { buildEditorUrl } from './build-url'
import { EditorSettings } from './editor-settings'

/**
 * @returns A function to open the current location in your preferred editor
 */
export const useOpenCurrentUrlInEditor = (): ((editorSettings: EditorSettings, sourcegraphURL: string) => void) => {
    const location = useLocation()
    return useCallback(
        (editorSettings: EditorSettings, sourcegraphURL: string) => {
            const { repoName, filePath, range } = parseBrowserRepoURL(location.pathname)
            const url = buildEditorUrl(
                `${repoName.split('/').pop() ?? ''}/${filePath}`,
                range,
                editorSettings,
                sourcegraphURL
            )
            window.open(url.toString(), '_blank')
        },
        [location.pathname]
    )
}
