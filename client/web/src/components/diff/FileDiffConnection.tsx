import React, { useCallback, useEffect, useMemo } from 'react'

import { EMPTY, from, combineLatest, ReplaySubject, BehaviorSubject, Observable } from 'rxjs'
import { concatMap, distinctUntilChanged, filter, map, mapTo, switchMap } from 'rxjs/operators'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'
import { Omit } from 'utility-types'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { ErrorLike, isDefined, isErrorLike, property } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { TextDocumentData, ViewerData, ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { FileSpec, RepoSpec, ResolvedRevisionSpec, RevisionSpec, toURIWithPath } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { FileDiffFields, Scalars } from '../../graphql-operations'
import { FilteredConnection, Connection } from '../FilteredConnection'

import { FileDiffNodeProps } from './FileDiffNode'

class FilteredFileDiffConnection extends FilteredConnection<FileDiffFields, NodeComponentProps> {}

type NodeComponentProps = Omit<FileDiffNodeProps, 'node'>

type FileDiffConnectionProps = FilteredFileDiffConnection['props']

// - Consumers of FileDiffConnection pass the minimum general extension data required for any file in the diff
// - FileDiffConnection, which is kind of an "extensions adapter layer" between general
// diff consumers and FilteredConnection, notifies the extension host of text documents and viewers,
// and adds `observeViewerId` to ExtensionInfo prop so diff node components can wait for extensions
// to know about viewers before hoverifying and listening for decorations (which depend on viewerId)
// - FileDiffNode also passes filePath for the diff part to the FileDiffHunks component, which uses
// the file path to construct uri to observe viewerId for that diff part.

/**
 * Information needed to apply extensions (hovers, decorations, ...) on the diff.
 * If undefined, extensions will not be applied on this diff.
 */
export type ExtensionInfo<ExtraExtensionData extends object = {}, ExtraPartData extends object = {}> = {
    /** The base repository, revision, and extra data (e.g. file). */
    base: PartInfo<ExtraPartData>

    /** The head repository, revision, and extra data (e.g. file). */
    head: PartInfo<ExtraPartData>
    hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>

    extensionsController: ExtensionsController
} & ExtraExtensionData

export type PartInfo<ExtraData extends object = {}> = {
    repoName: string
    repoID: Scalars['ID']
    revision: string
    commitID: string
} & ExtraData

/**
 * Displays a list of file diffs.
 */
export const FileDiffConnection: React.FunctionComponent<React.PropsWithChildren<FileDiffConnectionProps>> = props => {
    const { observeViewerId, setViewerIds, getCurrentViewerIdByUri } = useMemo(() => {
        const viewerIdByUris = new BehaviorSubject<Map<string, ViewerId | undefined>>(new Map())

        return {
            observeViewerId: (uri: string): Observable<ViewerId | undefined> =>
                viewerIdByUris.pipe(
                    map(viewerIdByUri => viewerIdByUri.get(uri)),
                    distinctUntilChanged()
                ),
            setViewerIds: async (
                projection: (
                    viewerIdByUri: Map<string, ViewerId | undefined>
                ) => Promise<Map<string, ViewerId | undefined>>
            ): Promise<void> => {
                viewerIdByUris.next(await projection(viewerIdByUris.value))
            },
            getCurrentViewerIdByUri: () => viewerIdByUris.value,
        }
    }, [])

    const extensionInfo = useMemo(() => props.nodeComponentProps?.extensionInfo, [
        props.nodeComponentProps?.extensionInfo,
    ])

    const extensionInfoChanges = useMemo(() => new ReplaySubject<ExtensionInfo | undefined>(1), [])
    useDeepCompareEffectNoCheck(() => {
        extensionInfoChanges.next(props.nodeComponentProps?.extensionInfo)
        // Use `useDeepCompareEffectNoCheck` since extensionInfo can be undefined
    }, [props.nodeComponentProps?.extensionInfo])

    // On unmount, get rid of all viewers
    useEffect(
        () => () => {
            const viewerIdByUri = getCurrentViewerIdByUri()
            extensionInfo?.extensionsController.extHostAPI
                .then(extensionHostAPI =>
                    Promise.all(
                        [...viewerIdByUri.values()]
                            .filter(isDefined)
                            .map(viewerId => extensionHostAPI.removeViewer(viewerId))
                    )
                )
                .catch(error => {
                    console.error('Error removing viewers from extension host', error)
                })
        },
        [getCurrentViewerIdByUri, extensionInfo?.extensionsController]
    )

    const diffsUpdates = useMemo(() => new ReplaySubject<Connection<FileDiffFields> | ErrorLike | undefined>(1), [])
    const nextDiffsUpdate: FileDiffConnectionProps['onUpdate'] = useCallback(
        fileDiffsOrError => diffsUpdates.next(fileDiffsOrError),
        [diffsUpdates]
    )

    // Add/remove viewers and documents on update
    useObservable(
        useMemo(
            () =>
                extensionInfoChanges.pipe(
                    filter(isDefined),
                    switchMap(extensionInfo =>
                        combineLatest([diffsUpdates, from(extensionInfo.extensionsController.extHostAPI)]).pipe(
                            concatMap(([fileDiffsOrError, extensionHostAPI]) => {
                                if (!fileDiffsOrError || isErrorLike(fileDiffsOrError)) {
                                    return EMPTY
                                }

                                // TODO(sqs): This reports to extensions that these files are empty. This is wrong, but we don't have any
                                // easy way to get the files' full contents here (and doing so would be very slow). Improve the extension
                                // API's support for diffs.
                                const dummyText = ''

                                // 1: Collect all base and head into uri -> initData (both document and viewer)
                                // 2: Iterate over existing uris w/ viewerIds, if not included in new record, schedule for removal.
                                // 3: Add all new text documents
                                // 4: Remove old viewers
                                // 5: Add new viewers
                                return from(
                                    setViewerIds(async currentViewerIdByUri => {
                                        try {
                                            const textDocumentDataByUri: Record<string, TextDocumentData> = {}
                                            const viewerDataByUri: Record<string, ViewerData> = {}
                                            const newUris = new Set<string>()

                                            for (const fileDiff of fileDiffsOrError.nodes) {
                                                // Base
                                                if (fileDiff.oldPath) {
                                                    const uri = toURIWithPath({
                                                        filePath: fileDiff.oldPath,
                                                        commitID: extensionInfo.base.commitID,
                                                        repoName: extensionInfo.base.repoName,
                                                    })
                                                    newUris.add(uri)

                                                    if (!currentViewerIdByUri.has(uri)) {
                                                        textDocumentDataByUri[uri] = {
                                                            uri,
                                                            languageId: getModeFromPath(fileDiff.oldPath),
                                                            text: dummyText,
                                                        }
                                                        viewerDataByUri[uri] = {
                                                            type: 'CodeEditor',
                                                            resource: uri,
                                                            selections: [],
                                                            isActive: true,
                                                        }
                                                    }
                                                }

                                                // Head
                                                if (fileDiff.newPath) {
                                                    const uri = toURIWithPath({
                                                        filePath: fileDiff.newPath,
                                                        commitID: extensionInfo.head.commitID,
                                                        repoName: extensionInfo.head.repoName,
                                                    })
                                                    newUris.add(uri)

                                                    if (!currentViewerIdByUri.has(uri)) {
                                                        textDocumentDataByUri[uri] = {
                                                            uri,
                                                            languageId: getModeFromPath(fileDiff.newPath),
                                                            text: dummyText,
                                                        }
                                                        viewerDataByUri[uri] = {
                                                            type: 'CodeEditor',
                                                            resource: uri,
                                                            selections: [],
                                                            isActive: true,
                                                        }
                                                    }
                                                }
                                            }

                                            // Determine viewers to remove
                                            const viewersToRemove = [...currentViewerIdByUri]
                                                .filter(([uri]) => !newUris.has(uri))
                                                .map(([uri, viewerId]) => ({ uri, viewerId }))
                                                .filter(property('viewerId', isDefined))

                                            // Register all text documents. Do this before registering viewers since they depend on documents.
                                            await Promise.all(
                                                Object.values(textDocumentDataByUri).map(textDocumentData =>
                                                    extensionHostAPI.addTextDocumentIfNotExists(textDocumentData)
                                                )
                                            )

                                            // Register new viewers with extension host
                                            const viewerIdsWithUri = await Promise.all(
                                                Object.values(viewerDataByUri).map(viewerData =>
                                                    extensionHostAPI
                                                        .addViewerIfNotExists(viewerData)
                                                        .then(viewerId => ({ uri: viewerData.resource, viewerId }))
                                                )
                                            )

                                            // Remove unused viewers from extension host
                                            await Promise.all(
                                                viewersToRemove.map(({ viewerId }) =>
                                                    extensionHostAPI.removeViewer(viewerId)
                                                )
                                            )

                                            // Update viewerId map (for diff components)
                                            for (const removedViewer of viewersToRemove) {
                                                currentViewerIdByUri.delete(removedViewer.uri)
                                            }
                                            for (const { uri, viewerId } of viewerIdsWithUri) {
                                                currentViewerIdByUri.set(uri, viewerId)
                                            }

                                            return new Map([...currentViewerIdByUri])
                                        } catch (error) {
                                            console.error('Error syncing documents/viewers with extension host', error)
                                            return new Map([...currentViewerIdByUri])
                                        }
                                    })
                                )
                            }),
                            mapTo(undefined)
                        )
                    )
                ),
            [extensionInfoChanges, diffsUpdates, setViewerIds]
        )
    )

    return (
        <FilteredFileDiffConnection
            {...props}
            nodeComponentProps={
                props.nodeComponentProps
                    ? {
                          ...props.nodeComponentProps,
                          extensionInfo: extensionInfo ? { ...extensionInfo, observeViewerId } : undefined,
                      }
                    : undefined
            }
            onUpdate={nextDiffsUpdate}
        />
    )
}
