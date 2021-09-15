import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useRef } from 'react'
import { from, Observable, ReplaySubject, Subscription } from 'rxjs'
import { map, mapTo, switchMap, tap } from 'rxjs/operators'

import { useBuiltinPanelViews } from '@sourcegraph/branded/src/components/panel/Panel'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import * as clientType from '@sourcegraph/extension-api-types'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ReferenceParameters, TextDocumentPositionParameters } from '@sourcegraph/shared/src/api/protocol'
import { Activation, ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { AbsoluteRepoFile, ModeSpec, parseQueryAndHash, UIPositionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { RepoRevisionSidebarCommits } from '../../RepoRevisionSidebarCommits'

interface Props extends AbsoluteRepoFile, ModeSpec, ExtensionsControllerProps, ActivationProps {
    location: H.Location
    history: H.History
    repoID: Scalars['ID']
    repoName: string
    commitID: string
}

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'impl' | 'typedef' | 'history'

/** The subject (what the contextual information refers to). */
interface PanelSubject extends AbsoluteRepoFile, ModeSpec, Partial<UIPositionSpec> {
    repoID: string

    /**
     * Include the full URI fragment here because it represents the state of panels, and we want
     * panels to be re-rendered when this state changes.
     */
    hash: string

    history: H.History
    location: H.Location
}

/**
 * A React hook that registers panel views for the blob.
 */
export function useBlobPanelViews({
    extensionsController,
    activation,
    repoName,
    commitID,
    revision,
    mode,
    filePath,
    repoID,
    location,
    history,
}: Props): void {
    const subscriptions = useMemo(() => new Subscription(), [])

    // Activation props are not stable
    const activationReference = useRef<Activation | undefined>(activation)
    activationReference.current = activation

    // Keep active code editor position subscription active to prevent empty loading state
    // (for main thread -> ext host -> main thread roundtrip for editor position)
    const activeCodeEditorPositions = useMemo(() => new ReplaySubject<TextDocumentPositionParameters | null>(1), [])
    useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getActiveCodeEditorPosition())),
                    tap(parameters => activeCodeEditorPositions.next(parameters)),
                    mapTo(undefined)
                ),
            [activeCodeEditorPositions, extensionsController]
        )
    )

    // Creates source for definition and reference panels
    const createLocationProvider = useCallback(
        <P extends TextDocumentPositionParameters>(
            id: string,
            title: string,
            priority: number,
            provideLocations: (parameters: P) => Observable<MaybeLoadingResult<clientType.Location[]>>,
            extraParameters?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParameters>>
        ) =>
            activeCodeEditorPositions.pipe(
                map(textDocumentPositionParameters => {
                    if (!textDocumentPositionParameters) {
                        return null
                    }

                    return {
                        title,
                        content: '',
                        priority,

                        // This disable directive is necessary because TypeScript is not yet smart
                        // enough to know that (typeof params & typeof extraParams) is P.
                        //
                        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                        locationProvider: provideLocations({
                            ...textDocumentPositionParameters,
                            ...extraParameters,
                        } as P).pipe(
                            tap(({ result: locations }) => {
                                if (activationReference.current && id === 'references' && locations.length > 0) {
                                    activationReference.current.update({ FoundReferences: true })
                                }
                            })
                        ),
                    }
                })
            ),
        [activeCodeEditorPositions]
    )

    // Source for history panel
    const panelSubject = useMemo(() => {
        const parsedHash = parseQueryAndHash(location.search, location.hash)
        return {
            repoID,
            repoName,
            commitID,
            revision,
            filePath,
            mode,
            position:
                parsedHash.line !== undefined
                    ? { line: parsedHash.line, character: parsedHash.character || 0 }
                    : undefined,
            hash: location.hash,
            history,
            location,
        }
    }, [commitID, filePath, history, location, mode, repoID, repoName, revision])

    const panelSubjectChanges = useMemo(() => new ReplaySubject<PanelSubject>(1), [])
    useEffect(() => {
        panelSubjectChanges.next(panelSubject)
    }, [panelSubject, panelSubjectChanges])

    useBuiltinPanelViews(
        useMemo(
            () => [
                {
                    id: 'history',
                    provider: panelSubjectChanges.pipe(
                        map(({ repoID, revision, filePath, history, location }) => ({
                            title: 'History',
                            content: '',
                            priority: 150,
                            locationProvider: undefined,
                            reactElement: (
                                <RepoRevisionSidebarCommits
                                    key="commits"
                                    repoID={repoID}
                                    revision={revision}
                                    filePath={filePath}
                                    history={history}
                                    location={location}
                                />
                            ),
                        }))
                    ),
                },
                {
                    id: 'def',
                    provider: createLocationProvider('def', 'Definition', 190, parameters =>
                        from(extensionsController.extHostAPI).pipe(
                            switchMap(extensionHostAPI =>
                                wrapRemoteObservable(extensionHostAPI.getDefinition(parameters))
                            )
                        )
                    ),
                },
                {
                    id: 'references',
                    provider: createLocationProvider<ReferenceParameters>('references', 'References', 180, parameters =>
                        from(extensionsController.extHostAPI).pipe(
                            switchMap(extensionHostAPI =>
                                wrapRemoteObservable(
                                    extensionHostAPI.getReferences(parameters, { includeDeclaration: false })
                                )
                            )
                        )
                    ),
                },
            ],
            [createLocationProvider, extensionsController.extHostAPI, panelSubjectChanges]
        )
    )

    useEffect(() => () => subscriptions.unsubscribe(), [subscriptions])
}
