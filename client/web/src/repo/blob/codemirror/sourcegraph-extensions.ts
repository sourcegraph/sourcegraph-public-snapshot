/**
 * This file contains CodeMirror extensions to integrate with Sourcegraph
 * extensions.
 *
 * This integration is done in various ways, see the specific extensions for
 * more information.
 */

import { Extension } from '@codemirror/state'
import { EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { Remote } from 'comlink'
import { combineLatest, EMPTY, from, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, shareReplay, switchMap } from 'rxjs/operators'

import { DocumentHighlight, emitLoading, LOADER_DELAY, MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, LineOrPositionOrRange, logger, lprToSelectionsZeroIndexed } from '@sourcegraph/common'
import { Position } from '@sourcegraph/extension-api-types'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { RequiredExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { toURIWithPath, UIPositionSpec } from '@sourcegraph/shared/src/util/url'

import { getHover } from '../../../backend/features'
import { BlobInfo } from '../Blob'

import { documentHighlightsSource } from './document-highlights'
import { hovercardSource } from './hovercard'
import { SelectedLineRange, selectedLines } from './linenumbers'

import { blobPropsFacet } from '.'

/**
 * Context holds all the information needed for CodeMirror extensions to
 * communicate with the extensions host.
 */
interface Context {
    viewerId: ViewerId
    extensionsController: RequiredExtensionsControllerProps['extensionsController']
    extensionHostAPI: Remote<FlatExtensionHostAPI>
    blobInfo: BlobInfo
}

/**
 * Enables integration with Sourcegraph extensions:
 * - Document highlights
 * - Hovercards (partially)
 * - Selection updates
 * - Reference panel warmup
 */
export function sourcegraphExtensions({
    blobInfo,
    initialSelection,
    extensionsController,
    enableSelectionDrivenCodeNavigation,
}: {
    blobInfo: BlobInfo
    initialSelection: LineOrPositionOrRange
    extensionsController: RequiredExtensionsControllerProps['extensionsController']
    enableSelectionDrivenCodeNavigation?: boolean
}): Extension {
    const subscriptions = new Subscription()

    // Initialize document and viewer as early as possible and make context
    // available as observable
    const contextObservable = from(
        extensionsController.extHostAPI.then(async extensionHostAPI => {
            const uri = toURIWithPath(blobInfo)

            const [, viewerId] = await Promise.all([
                // This call should be made before adding viewer, but since
                // messages to web worker are handled in order, we can use Promise.all
                extensionHostAPI.addTextDocumentIfNotExists({
                    uri,
                    languageId: blobInfo.mode,
                    text: blobInfo.content,
                }),
                extensionHostAPI.addViewerIfNotExists({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    selections: lprToSelectionsZeroIndexed(initialSelection),
                    isActive: true,
                }),
            ])

            return [viewerId, extensionHostAPI] as [ViewerId, Remote<FlatExtensionHostAPI>]
        })
    ).pipe(
        catchError(() => {
            logger.error('Unable to initialize extensions context')
            return EMPTY
        }),
        map(([viewerId, extensionHostAPI]) => {
            subscriptions.add(() => {
                extensionHostAPI
                    .removeViewer(viewerId)
                    .catch(error => logger.error('Error removing viewer from extension host', error))
            })

            return {
                blobInfo,
                viewerId,
                extensionHostAPI,
                extensionsController,
            }
        }),
        shareReplay(1)
    )

    return [
        // This view plugin is used to have a way to cleanup any resources via the
        // `destroy` method.
        ViewPlugin.define(() => ({
            destroy() {
                subscriptions.unsubscribe()
            },
        })),
        // This needs to come before document highlights so that the hovered
        // token is highlighted differently. Hovercard datasource is disabled
        // when selection-driven navigation is enabled because it reimplements
        // hover logic in the file 'token-selection/hover.ts'.
        enableSelectionDrivenCodeNavigation ? [] : hovercardDataSource(contextObservable),
        enableSelectionDrivenCodeNavigation ? [] : documentHighlightsDataSource(contextObservable),
        ViewPlugin.define(() => new SelectionManager(contextObservable)),
        ViewPlugin.define(() => new WarmupReferencesManager(contextObservable)),
    ]
}

//
// Document highlights
//

/**
 * documentHighlightsDataSource registers a callback function for retrieving
 * document highlight information.
 * See {@link DocumentHighlightsDataSource} and {@link documentHighlightsSource}.
 */
function documentHighlightsDataSource(context: Observable<Context>): Extension {
    return documentHighlightsSource.of(
        (position: Position): Observable<DocumentHighlight[]> =>
            combineLatest([context, of(position)]).pipe(
                switchMap(([context, position]) =>
                    wrapRemoteObservable(
                        context.extensionHostAPI.getDocumentHighlights({
                            textDocument: {
                                uri: toURIWithPath(context.blobInfo),
                            },
                            position: {
                                character: position.character - 1,
                                line: position.line - 1,
                            },
                        })
                    )
                )
            )
    )
}

//
// Text document decorations
//

//
// Selection change notifier
//

/**
 * The selection manager listens to CodeMirror selection changes and sends them
 * to the extensions host.
 */
class SelectionManager implements PluginValue {
    private nextSelection: Subject<SelectedLineRange> = new Subject()
    private subscription = new Subscription()

    constructor(context: Observable<Context>) {
        this.subscription = combineLatest([context, this.nextSelection]).subscribe(([context, selection]) => {
            // Used to convert SelectedLineRange type to a valid LineOrPositionOrRange type to keep TypeScript happy
            let lprSelection: LineOrPositionOrRange = {}
            if (selection) {
                lprSelection =
                    selection.line !== selection.endLine
                        ? { line: selection.line, endLine: selection.endLine }
                        : { line: selection.line, character: selection.character }
            }
            context.extensionHostAPI
                .setEditorSelections(context.viewerId, lprToSelectionsZeroIndexed(lprSelection))
                .catch(error => logger.error('Error updating editor selections on extension host', error))
        })
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }

    public update(update: ViewUpdate): void {
        if (update.state.field(selectedLines) !== update.startState.field(selectedLines)) {
            this.nextSelection.next(update.state.field(selectedLines))
        }
    }
}

//
// Hovercards
//

/**
 * hovercardDataSource uses the {@link hovercardSource} facet to provide a
 * callback function for querying the extension API for hover data.
 */
function hovercardDataSource(context: Observable<Context>): Extension {
    return hovercardSource.of(
        (
            view: EditorView,
            position: UIPositionSpec['position']
        ): Observable<Pick<HoverOverlayBaseProps, 'hoverOrError' | 'actionsOrError'>> =>
            context.pipe(
                filter((context): context is Context => context !== null),
                switchMap(context => {
                    const hoverContext = {
                        commitID: context.blobInfo.commitID,
                        revision: context.blobInfo.revision,
                        filePath: context.blobInfo.filePath,
                        repoName: context.blobInfo.repoName,
                    }
                    const { extensionsController, platformContext } = view.state.facet(blobPropsFacet)

                    return combineLatest([
                        getHover({ ...hoverContext, position }, { extensionsController }).pipe(
                            catchError((error): [MaybeLoadingResult<ErrorLike>] => [
                                { isLoading: false, result: asError(error) },
                            ]),
                            emitLoading<HoverOverlayBaseProps['hoverOrError'] | ErrorLike, null>(LOADER_DELAY, null)
                        ),
                        getHoverActions(
                            { extensionsController, platformContext },
                            {
                                ...hoverContext,
                                ...position,
                            }
                        ),
                    ])
                }),
                map(([hoverResult, actionsResult]) => ({
                    hoverOrError: hoverResult,
                    actionsOrError: actionsResult,
                }))
            )
    )
}

class WarmupReferencesManager implements PluginValue {
    private subscription: Subscription

    constructor(context: Observable<Context>) {
        this.subscription = context
            .pipe(switchMap(context => haveInitialExtensionsLoaded(context.extensionsController.extHostAPI)))
            .subscribe()
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }
}
