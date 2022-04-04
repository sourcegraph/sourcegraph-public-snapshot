import { useMemo, useEffect } from 'react'

import { ReplaySubject } from 'rxjs'

import { ViewerId } from '../api/viewerTypes'
import { ExtensionsControllerProps } from '../extensions/controller'
import { HoverContext } from '../hover/HoverOverlay'
import { getModeFromPath } from '../languages'

import { toURIWithPath } from './url'

interface UseCodeIntelViewerUpdatesProps extends ExtensionsControllerProps<'extHostAPI'> {
    repositoryName: string
    filePath: string
    revision: string | undefined
}

type ViewerUpdate = { viewerId: ViewerId } & HoverContext

export function useCodeIntelViewerUpdates(props?: UseCodeIntelViewerUpdatesProps): ReplaySubject<ViewerUpdate> {
    // Inform the extension host about the file (if we have code to render). CodeExcerpt will call `hoverifier.hoverify`.
    const viewerUpdates = useMemo(() => new ReplaySubject<ViewerUpdate>(1), [])
    useEffect(() => {
        if (!props) {
            return
        }

        const { extensionsController, repositoryName, revision, filePath } = props

        let previousViewerId: ViewerId | undefined
        const commitID = revision || 'HEAD'
        const uri = toURIWithPath({ repoName: repositoryName, filePath, commitID })
        const languageId = getModeFromPath(filePath)
        // HACK: code intel extensions don't depend on the `text` field.
        // Fix to support other hover extensions on search results (likely too expensive).
        const text = ''

        extensionsController.extHostAPI
            .then(extensionHostAPI =>
                Promise.all([
                    // This call should be made before adding viewer, but since
                    // messages to web worker are handled in order, we can use Promise.all
                    extensionHostAPI.addTextDocumentIfNotExists({ uri, languageId, text }),
                    extensionHostAPI.addViewerIfNotExists({
                        type: 'CodeEditor' as const,
                        resource: uri,
                        selections: [],
                        isActive: true,
                    }),
                ])
            )
            .then(([, viewerId]) => {
                previousViewerId = viewerId
                viewerUpdates.next({
                    viewerId,
                    repoName: repositoryName,
                    revision: commitID,
                    commitID,
                    filePath,
                })
            })
            .catch(error => {
                console.error('Extension host API error', error)
            })

        return () => {
            // Remove from extension host
            extensionsController.extHostAPI
                .then(extensionHostAPI => previousViewerId && extensionHostAPI.removeViewer(previousViewerId))
                .catch(error => console.error('Error removing viewer from extension host', error))
        }
    }, [props, viewerUpdates])

    return viewerUpdates
}
