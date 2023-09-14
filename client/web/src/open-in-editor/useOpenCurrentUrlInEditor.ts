import { useCallback } from 'react'

import { parseBrowserRepoURL } from '../util/url'

import { buildEditorUrl, buildRepoBaseNameAndPath } from './build-url'
import type { EditorSettings } from './editor-settings'

/**
 * @returns A function to open the current location in your preferred editor
 */
export const useOpenCurrentUrlInEditor = (): ((
    editorSettings: EditorSettings | undefined,
    externalServiceType: string | undefined,
    sourcegraphURL: string,
    editorIndex?: number
) => void) =>
    useCallback(
        (
            editorSettings: EditorSettings | undefined,
            externalServiceType: string | undefined,
            sourcegraphURL: string,
            editorIndex = 0
        ) => {
            const { repoName, filePath, position, range } = parseBrowserRepoURL(window.location.href)
            const start = position || range?.start
            const url = buildEditorUrl(
                buildRepoBaseNameAndPath(repoName, externalServiceType, filePath),
                start,
                editorSettings,
                sourcegraphURL,
                editorIndex
            )
            window.open(url.toString(), '_blank')
        },
        []
    )
