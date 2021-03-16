import { findPositionsFromEvents } from '@sourcegraph/codeintellify'
import * as H from 'history'
import React, { useCallback, useMemo, useState, useEffect } from 'react'
import { combineLatest, from, NEVER, Observable, of, ReplaySubject, Subscription } from 'rxjs'
import { filter, first, switchMap, tap } from 'rxjs/operators'
import { DecorationMapByLine, groupDecorationsByLine } from '../../../../shared/src/api/extension/api/decorations'
import { isDefined, property } from '../../../../shared/src/util/types'
import { toURIWithPath } from '../../../../shared/src/util/url'
import { ThemeProps } from '../../../../shared/src/theme'
import { DiffHunk } from './DiffHunk'
import { diffDomFunctions } from '../../repo/compare/dom-functions'
import { FileDiffFields } from '../../graphql-operations'
import { ViewerId } from '../../../../shared/src/api/viewerTypes'
import { ExtensionInfo } from './FileDiffConnection'
import { wrapRemoteObservable } from '../../../../shared/src/api/client/api/common'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'
import { useObservable } from '../../../../shared/src/util/useObservable'

export interface FileHunksProps extends ThemeProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string

    /**
     * Information needed to apply extensions (hovers, decorations, ...) on the diff.
     * If undefined, extensions will not be applied on this diff.
     */
    extensionInfo?: ExtensionInfo<
        {
            observeViewerId?: (uri: string) => Observable<ViewerId | undefined>
        },
        {
            /**
             * `null` if the file does not exist in this diff part.
             */
            filePath: string | null
        }
    >

    /** The file's hunks. */
    hunks: FileDiffFields['hunks']

    /** Whether to show line numbers. */
    lineNumbers: boolean

    className: string
    location: H.Location
    history: H.History
    /** Reflect selected line in url */
    persistLines?: boolean
}

/** Displays hunks in a unified file diff. */
export const FileDiffHunks: React.FunctionComponent<FileHunksProps> = ({
    className,
    fileDiffAnchor,
    history,
    hunks,
    isLightTheme,
    lineNumbers,
    location,
    extensionInfo,
    persistLines,
}) => {
    /**
     * Decorations for the file at the two revisions of the diff
     */
    const [decorations, setDecorations] = useState<Record<'head' | 'base', DecorationMapByLine>>({
        head: new Map(),
        base: new Map(),
    })

    /** Emits whenever the ref callback for the code element is called */
    const codeElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextCodeElement = useCallback((element: HTMLElement | null): void => codeElements.next(element), [
        codeElements,
    ])

    /** Emits whenever the ref callback for the blob element is called */
    const blobElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextBlobElement = useCallback((element: HTMLElement | null): void => blobElements.next(element), [
        blobElements,
    ])

    const extensionInfoChanges = useMemo(() => new ReplaySubject<FileHunksProps['extensionInfo'] | undefined>(1), [])
    useDeepCompareEffectNoCheck(() => {
        extensionInfoChanges.next(extensionInfo)
        // Use `useDeepCompareEffectNoCheck` since extensionInfo can be undefined
    }, [extensionInfo])

    // Listen for line decorations from extensions
    useObservable(
        useMemo(
            () =>
                extensionInfoChanges.pipe(
                    filter(isDefined),
                    filter(property('observeViewerId', isDefined)),
                    switchMap(extensionInfo => {
                        const baseUri = extensionInfo.base.filePath
                            ? toURIWithPath({
                                  repoName: extensionInfo.base.repoName,
                                  commitID: extensionInfo.base.commitID,
                                  filePath: extensionInfo.base.filePath,
                              })
                            : null
                        const baseViewerIds = baseUri ? extensionInfo.observeViewerId(baseUri) : of(null)

                        const headUri = extensionInfo.head.filePath
                            ? toURIWithPath({
                                  repoName: extensionInfo.head.repoName,
                                  commitID: extensionInfo.head.commitID,
                                  filePath: extensionInfo.head.filePath,
                              })
                            : null
                        const headViewerIds = headUri ? extensionInfo.observeViewerId(headUri) : of(null)

                        return combineLatest([
                            baseViewerIds,
                            headViewerIds,
                            from(extensionInfo.extensionsController.extHostAPI),
                        ]).pipe(
                            switchMap(([baseViewerId, headViewerId, extensionHostAPI]) =>
                                combineLatest([
                                    baseViewerId
                                        ? wrapRemoteObservable(extensionHostAPI.getTextDecorations(baseViewerId))
                                        : of(null),
                                    headViewerId
                                        ? wrapRemoteObservable(extensionHostAPI.getTextDecorations(headViewerId))
                                        : of(null),
                                ])
                            )
                        )
                    }),
                    tap(([baseDecorations, headDecorations]) => {
                        setDecorations({
                            base: groupDecorationsByLine(baseDecorations),
                            head: groupDecorationsByLine(headDecorations),
                        })
                    })
                ),
            [extensionInfoChanges]
        )
    )

    // Hoverify
    useEffect(() => {
        const subscription = new Subscription()

        let hoverSubscription: Subscription
        subscription.add(
            extensionInfoChanges.pipe(filter(isDefined), first()).subscribe(extensionInfo => {
                hoverSubscription?.unsubscribe()

                hoverSubscription = extensionInfo.hoverifier.hoverify({
                    dom: diffDomFunctions,
                    positionEvents: codeElements.pipe(
                        filter(isDefined),
                        findPositionsFromEvents({ domFunctions: diffDomFunctions })
                    ),
                    positionJumps: NEVER, // TODO support diff URLs
                    resolveContext: hoveredToken => {
                        // if part is undefined, it doesn't matter whether we chose head or base, the line stayed the same
                        const { repoName, revision, filePath, commitID } = extensionInfo[hoveredToken.part || 'head']

                        // If a hover or go-to-definition was invoked on this part, we know the file path must exist
                        return { repoName, filePath: filePath!, revision, commitID }
                    },
                })
                subscription.add(hoverSubscription)
            })
        )

        return () => subscription.unsubscribe()
    }, [codeElements, extensionInfoChanges])

    return (
        <div className={`file-diff-hunks ${className}`} ref={nextBlobElement}>
            {hunks.length === 0 ? (
                <div className="text-muted m-2">No changes</div>
            ) : (
                <div className="file-diff-hunks__container" ref={nextCodeElement}>
                    <table className="file-diff-hunks__table">
                        {lineNumbers && (
                            <colgroup>
                                <col width="40" />
                                <col width="40" />
                                <col />
                            </colgroup>
                        )}
                        <tbody>
                            {hunks.map((hunk, index) => (
                                <DiffHunk
                                    fileDiffAnchor={fileDiffAnchor}
                                    history={history}
                                    isLightTheme={isLightTheme}
                                    lineNumbers={lineNumbers}
                                    location={location}
                                    persistLines={persistLines}
                                    key={index}
                                    hunk={hunk}
                                    decorations={decorations}
                                />
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}
