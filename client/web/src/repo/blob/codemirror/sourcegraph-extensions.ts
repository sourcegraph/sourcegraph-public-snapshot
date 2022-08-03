/**
 * This file contains CodeMirror extensions to integrate with Sourcegraph
 * extensions.
 *
 * This integration is done in various ways, see the specific extensions for
 * more information.
 */
import React from 'react'

import { Extension, Facet, StateEffect, StateEffectType, StateField } from '@codemirror/state'
import { EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { Remote } from 'comlink'
import { createRoot, Root } from 'react-dom/client'
import { combineLatest, Observable, of, ReplaySubject, Subject, Subscription } from 'rxjs'
import { filter, map, catchError, switchMap, distinctUntilChanged, startWith } from 'rxjs/operators'
import { TextDocumentDecorationType } from 'sourcegraph'

import {
    DocumentHighlight,
    LineOrPositionOrRange,
    LOADER_DELAY,
    MaybeLoadingResult,
    emitLoading,
} from '@sourcegraph/codeintellify'
import { asError, ErrorLike, lprToSelectionsZeroIndexed } from '@sourcegraph/common'
import { Position, TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { StatusBarItemWithKey } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverOverlayBaseProps } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { toURIWithPath, UIPositionSpec } from '@sourcegraph/shared/src/util/url'

import { getHover } from '../../../backend/features'
import { StatusBar } from '../../../extensions/components/StatusBar'
import { BlobInfo, BlobProps } from '../Blob'

import { showTextDocumentDecorations } from './decorations'
import { documentHighlightsSource } from './document-highlights'
import { hovercardSource } from './hovercard'
import { Container } from './react-interop'

import { blobPropsFacet } from '.'

import blobStyles from '../Blob.module.scss'

/**
 * Context holds all the information needed for CodeMirror extensions to
 * communicate with the extensions host.
 */
interface Context {
    viewerId: ViewerId
    extensionsController: ExtensionsControllerProps['extensionsController']
    extensionHostAPI: Remote<FlatExtensionHostAPI>
    blobInfo: BlobInfo
    subscriptions: Subscription
}

/**
 * Enables integration with Sourcegraph extensions:
 * - Document highlights
 * - Hovercards (partially)
 * - Text document decorations
 * - Selection updates
 * - Status bar
 */
export function sourcegraphExtensions({
    blobInfo,
    extensionsController,
}: {
    blobInfo: BlobInfo
    extensionsController: ExtensionsControllerProps['extensionsController']
}): Extension {
    const context = extensionsController.extHostAPI.then(async extensionHostAPI => {
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
                // TODO: set initial selection from selected lines
                selections: [], // lprToSelectionsZeroIndexed(initialPosition),
                isActive: true,
            }),
        ])

        return [viewerId, extensionHostAPI] as [ViewerId, Remote<FlatExtensionHostAPI>]
    })

    return [
        // A view plugin is used to have a way to cleanup any resources via the
        // `destroy` method.
        ViewPlugin.define(view => {
            context
                .then(([viewerId, extensionHostAPI]) => {
                    const subscriptions = new Subscription()

                    // Cleanup on navigation between/away from viewers
                    subscriptions.add(() => {
                        extensionHostAPI
                            .removeViewer(viewerId)
                            .catch(error => console.error('Error removing viewer from extension host', error))
                    })

                    view.dispatch({
                        effects: setContext.of({
                            blobInfo,
                            viewerId,
                            extensionHostAPI,
                            subscriptions,
                            extensionsController,
                        }),
                    })
                })
                .catch(() => console.error('Unable to initialize extensions context'))

            return {
                destroy() {
                    const context = view.state.field(sgExtensionsContextField, false)?.context
                    context?.subscriptions.unsubscribe()
                },
            }
        }),
        sgExtensionsContextField,
        // This needs to come before document highlights so that the hovered
        // token is highlighted differently
        hovercardDataSource(),
        documentHighlightsDataSource(),
        textDocumentDecorations(),
        updateSelection(),
        statusBar,
        warmupReferences,
    ]
}

/**
 * Only used by sgExtenionContext to initialize the context.
 */
const setContext = StateEffect.define<Context>()

/**
 * Stores the context which is necessary to integrate with Sourcegraph
 * extensions. Since it takes some time to initialize the connection to the
 * extension host, CodeMirror extensions can either use the observable stored in
 * this field or the {@link updateOnContextChange} facet to be notified when the
 * context becomes available.
 */
const sgExtensionsContextField = StateField.define<{ context: Context | null; contextObservable: Subject<Context> }>({
    create() {
        return {
            context: null,
            contextObservable: new ReplaySubject(1),
        }
    },
    compare(previous, next) {
        return previous.context === next.context
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setContext)) {
                value.contextObservable.next(effect.value)
                return { ...value, context: effect.value }
            }
        }
        return value
    },
})

/**
 * Facet that allows other extensions to get notified when the Sourcegraph
 * extensions context is initialized.
 */
const updateOnContextChange = Facet.define<
    ViewPlugin<{ setContext(context: Context | null): void }> | ((context: Context | null) => void)
>({
    enables: facet =>
        EditorView.updateListener.of(update => {
            const context = update.state.field(sgExtensionsContextField)
            if (update.startState.field(sgExtensionsContextField) !== context) {
                for (const listener of update.state.facet(facet)) {
                    if (listener instanceof ViewPlugin) {
                        update.view.plugin(listener)?.setContext(context.context)
                    } else {
                        listener(context.context)
                    }
                }
            }
        }),
})

//
// Document highlights
//

/**
 * Listen to CodeMirror events and generating decorations for document
 * highlights is done completely independently of Sourcegraph extensions. The
 * integration happens by registering a "data source" that the input extension
 * can query.
 * See {@link DocumentHighlightsDataSource} and {@link documentHighlights\Sources}.
 */
function documentHighlightsDataSource(): Extension {
    const nextContext: Subject<Context | null> = new ReplaySubject(1)
    const EMPTY: DocumentHighlight[] = []

    const createObservable = (position: Position): Observable<DocumentHighlight[]> =>
        combineLatest([nextContext, of(position)]).pipe(
            switchMap(([context, position]) =>
                context
                    ? wrapRemoteObservable(
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
                    : of(EMPTY)
            )
        )

    return [
        updateOnContextChange.of(context => nextContext.next(context)),
        documentHighlightsSource.of(createObservable),
    ]
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
class TextDecorationManager {
    private subscription: Subscription

    constructor(
        private readonly view: EditorView,
        private readonly setDecorations: StateEffectType<[TextDocumentDecorationType, TextDocumentDecoration[]][]>
    ) {
        this.subscription = this.view.state
            .field(sgExtensionsContextField)
            .contextObservable.pipe(
                switchMap(context =>
                    context
                        ? wrapRemoteObservable(context.extensionHostAPI.getTextDecorations(context.viewerId))
                        : of([])
                )
            )
            .subscribe(decorations => {
                view.dispatch({ effects: this.setDecorations.of(decorations) })
            })
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }
}

function textDocumentDecorations(): Extension {
    const [decorationsField, , setDecorations] = createUpdateableField<
        [TextDocumentDecorationType, TextDocumentDecoration[]][]
    >([], field => showTextDocumentDecorations.from(field))
    return [decorationsField, ViewPlugin.define(view => new TextDecorationManager(view, setDecorations))]
}

//
// Selection change notifier
//

/**
 * The selection manager listens to CodeMirror selection changes and sends them
 * to the extensions host. This extension listens directly to selection changes
 * because it's a simple task and introducing an integration point with a
 * separate CodeMirror extension would just make things more complicated.
 */
class SelectionManager {
    private context: Context | null = null
    private selectionSet = false

    constructor(private readonly view: EditorView) {}

    public update(viewUpdate: ViewUpdate): void {
        if (viewUpdate.selectionSet) {
            this.selectionSet = true
            this.updatePosition()
        }
    }

    public setContext(context: Context | null): void {
        this.context = context
        if (this.selectionSet) {
            this.updatePosition()
        }
    }

    private updatePosition(): void {
        if (this.context) {
            const selection = this.view.state.selection.main
            const fromLine = this.view.state.doc.lineAt(selection.from)
            const toLine = selection.from === selection.to ? fromLine : this.view.state.doc.lineAt(selection.to)

            const position: LineOrPositionOrRange = {
                line: fromLine.number,
                endLine: toLine.number,
            }

            this.context.extensionHostAPI
                .setEditorSelections(this.context.viewerId, lprToSelectionsZeroIndexed(position))
                .catch(error => console.error('Error updating editor selections on extension host', error))
        }
    }
}

function updateSelection(): Extension {
    return ViewPlugin.fromClass(SelectionManager, {
        provide: plugin => updateOnContextChange.of(plugin),
    })
}

//
// Hovercards
//

/**
 * hovercardDataSource uses the {@link hovercardSource} facet to provide a
 * callback function for querying the extension API for hover data.
 */
function hovercardDataSource(): Extension {
    const nextContext: Subject<Context | null> = new ReplaySubject(1)

    const createObservable = (
        view: EditorView,
        position: UIPositionSpec['position']
    ): Observable<Pick<HoverOverlayBaseProps, 'hoverOrError' | 'actionsOrError'>> =>
        nextContext.pipe(
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

    return [updateOnContextChange.of(context => nextContext.next(context)), hovercardSource.of(createObservable)]
}

//
// Status bar
//

/**
 * The status bar integration doesn't require to integrate with the input or output
 * capabilities of CodeMirror. It only attaches a container DOM element to the
 * editor's DOM and renders itself it that container.
 */
const statusBar = ViewPlugin.fromClass(
    class {
        private container: HTMLDivElement
        private reactRoot: Root
        private subscription: Subscription
        private nextProps = new Subject<BlobProps>()

        constructor(private readonly view: EditorView) {
            this.container = document.createElement('div')
            this.reactRoot = createRoot(this.container)
            const contextUpdates = this.view.state.field(sgExtensionsContextField).contextObservable

            const getStatusBarItems = (): Observable<'loading' | StatusBarItemWithKey[]> =>
                contextUpdates.pipe(
                    switchMap(context => {
                        if (!context) {
                            return of('loading' as const)
                        }

                        return wrapRemoteObservable(context.extensionHostAPI.getStatusBarItems(context.viewerId))
                    })
                )

            this.subscription = combineLatest([
                contextUpdates,
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

            this.view.dom.append(this.container)
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
)

const warmupReferences = ViewPlugin.fromClass(
    class {
        private nextContext: Subject<Context> = new Subject()
        private subscription: Subscription = new Subscription()

        constructor() {
            this.subscription.add(
                this.nextContext
                    .pipe(switchMap(context => haveInitialExtensionsLoaded(context.extensionsController.extHostAPI)))
                    .subscribe()
            )
        }

        public setContext(context: Context): void {
            this.nextContext.next(context)
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    },
    {
        provide: plugin => updateOnContextChange.of(plugin),
    }
)
