/**
 * This file contains CodeMirror extensions to integrate with Sourcegraph
 * extensions.
 *
 * This integration is done in various ways, see the specific extensions for
 * more information.
 */
import React from 'react'

import { Extension, StateEffectType } from '@codemirror/state'
import { EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { Remote } from 'comlink'
import { createRoot, Root } from 'react-dom/client'
import { combineLatest, EMPTY, from, Observable, of, Subject, Subscription } from 'rxjs'
import { filter, map, catchError, switchMap, distinctUntilChanged, startWith, shareReplay } from 'rxjs/operators'

import { DocumentHighlight, LOADER_DELAY, MaybeLoadingResult, emitLoading } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, LineOrPositionOrRange, lprToSelectionsZeroIndexed } from '@sourcegraph/common'
import { Position, TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { RequiredExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { toURIWithPath, UIPositionSpec } from '@sourcegraph/shared/src/util/url'

import { getHover } from '../../../backend/features'
import { StatusBar } from '../../../extensions/components/StatusBar'
import { BlobInfo, BlobProps } from '../Blob'

import { documentHighlightsSource } from './document-highlights'
import { showTextDocumentDecorations } from './extensions-decorations'
import { hovercardSource } from './hovercard'
import { SelectedLineRange, selectedLines } from './linenumbers'
import { Container } from './react-interop'

import { blobPropsFacet } from '.'

import blobStyles from '../Blob.module.scss'

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
 * - Text document decorations
 * - Selection updates
 * - Status bar
 * - Reference panel warmup
 */
export function sourcegraphExtensions({
    blobInfo,
    initialSelection,
    extensionsController,
    disableStatusBar,
    disableDecorations,
}: {
    blobInfo: BlobInfo
    initialSelection: LineOrPositionOrRange
    extensionsController: RequiredExtensionsControllerProps['extensionsController']
    disableStatusBar?: boolean
    disableDecorations?: boolean
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
            console.error('Unable to initialize extensions context')
            return EMPTY
        }),
        map(([viewerId, extensionHostAPI]) => {
            subscriptions.add(() => {
                extensionHostAPI
                    .removeViewer(viewerId)
                    .catch(error => console.error('Error removing viewer from extension host', error))
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
        // token is highlighted differently
        hovercardDataSource(contextObservable),
        documentHighlightsDataSource(contextObservable),
        disableDecorations ? [] : textDocumentDecorations(contextObservable),
        ViewPlugin.define(() => new SelectionManager(contextObservable)),
        disableStatusBar
            ? []
            : [ViewPlugin.define(view => new StatusBarManager(view, contextObservable)), bottomPadding],
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

/**
 * This integration doesn't require any input from CodeMirror. Rendering text
 * document decorations is done independently on the CodeMirror side.
 * TextDecorationManager manages the subscription to the extension host and uses
 * a state field to provide input values for the {@link showTextDocumentDecorations}
 * facet.
 */
class TextDecorationManager implements PluginValue {
    private subscription: Subscription

    constructor(
        view: EditorView,
        context: Observable<Context>,
        setDecorations: StateEffectType<TextDocumentDecoration[]>
    ) {
        this.subscription = context
            .pipe(
                switchMap(context =>
                    wrapRemoteObservable(context.extensionHostAPI.getTextDecorations(context.viewerId))
                )
            )
            .subscribe(decorations => {
                view.dispatch({ effects: setDecorations.of(decorations) })
            })
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }
}

function textDocumentDecorations(context: Observable<Context>): Extension {
    const [decorationsField, , setDecorations] = createUpdateableField<TextDocumentDecoration[]>([], field =>
        showTextDocumentDecorations.from(field)
    )
    return [decorationsField, ViewPlugin.define(view => new TextDecorationManager(view, context, setDecorations))]
}

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
            context.extensionHostAPI
                .setEditorSelections(context.viewerId, lprToSelectionsZeroIndexed(selection ?? {}))
                .catch(error => console.error('Error updating editor selections on extension host', error))
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

//
// Status bar
//

/**
 * The status bar integration doesn't require to integrate with the input or output
 * capabilities of CodeMirror. It only attaches a container DOM element to the
 * editor's DOM and renders itself it that container.
 */
class StatusBarManager implements PluginValue {
    private container: HTMLDivElement
    private reactRoot: Root
    private subscription: Subscription
    private nextProps = new Subject<BlobProps>()

    constructor(view: EditorView, context: Observable<Context>) {
        this.container = document.createElement('div')
        this.reactRoot = createRoot(this.container)

        const getStatusBarItems = (): Observable<'loading' | StatusBarItemWithKey[]> =>
            context.pipe(
                switchMap(context => {
                    if (!context) {
                        return of('loading' as const)
                    }

                    return wrapRemoteObservable(context.extensionHostAPI.getStatusBarItems(context.viewerId))
                })
            )

        this.subscription = combineLatest([
            context,
            this.nextProps.pipe(
                distinctUntilChanged(
                    (previous, next) => previous.location === next.location && previous.history === next.history
                ),
                startWith(view.state.facet(blobPropsFacet))
            ),
        ]).subscribe(([context, props]) => {
            this.reactRoot.render(
                React.createElement(
                    Container,
                    { history: props.history },
                    React.createElement(StatusBar, {
                        getStatusBarItems,
                        extensionsController: context.extensionsController,
                        uri: toURIWithPath(context.blobInfo),
                        location: props.location,
                        className: blobStyles.blobStatusBarBody,
                        statusBarRef: () => {},
                        hideWhileInitializing: true,
                        isBlobPage: true,
                    })
                )
            )
        })

        view.dom.append(this.container)
    }

    public update(update: ViewUpdate): void {
        this.nextProps.next(update.state.facet(blobPropsFacet))
    }

    public destroy(): void {
        this.subscription.unsubscribe()
        this.container.remove()

        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.reactRoot.unmount(), 0)
    }
}

// Ensures that the status bar doesn't conver any content
const bottomPadding = EditorView.theme({
    '.cm-content': {
        paddingBottom: 'calc(var(--blob-status-bar-height) + var(--blob-status-bar-vertical-gap))',
    },
})

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
