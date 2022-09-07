import { proxy } from 'comlink'
import { castArray, groupBy, identity, isEqual } from 'lodash'
import { combineLatest, concat, EMPTY, from, Observable, of, Subscribable, throwError } from 'rxjs'
import {
    catchError,
    debounceTime,
    defaultIfEmpty,
    distinctUntilChanged,
    map,
    mergeMap,
    switchMap,
} from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'

import {
    fromHoverMerged,
    TextDocumentIdentifier,
    ContributableViewContainer,
    TextDocumentPositionParameters,
} from '@sourcegraph/client-api'
import { LOADING, MaybeLoadingResult } from '@sourcegraph/codeintellify'
import {
    allOf,
    asError,
    combineLatestOrDefault,
    ErrorLike,
    isDefined,
    isExactly,
    isNot,
    property,
} from '@sourcegraph/common'
import * as clientType from '@sourcegraph/extension-api-types'
import { Context } from '@sourcegraph/template-parser'

import { getModeFromPath } from '../../languages'
import { parseRepoURI } from '../../util/url'
import { match } from '../client/types/textDocument'
import { FlatExtensionHostAPI } from '../contract'
import { ExtensionViewer, ViewerId, ViewerWithPartialModel } from '../viewerTypes'

import { ExtensionCodeEditor } from './api/codeEditor'
import { providerResultToObservable, ProxySubscribable, proxySubscribable } from './api/common'
import { computeContext, ContributionScope } from './api/context/context'
import {
    evaluateContributions,
    filterContributions,
    mergeContributions,
    parseContributionExpressions,
} from './api/contribution'
import { validateFileDecoration } from './api/decorations'
import { ExtensionDirectoryViewer } from './api/directoryViewer'
import { getInsightsViews } from './api/getInsightsViews'
import { ExtensionDocument } from './api/textDocument'
import { fromLocation, toPosition } from './api/types'
import { ExtensionWorkspaceRoot } from './api/workspaceRoot'
import { updateContext } from './extensionHost'
import { ExtensionHostState } from './extensionHostState'
import { addWithRollback } from './util'

export function createExtensionHostAPI(state: ExtensionHostState): FlatExtensionHostAPI {
    const getTextDocument = (uri: string): ExtensionDocument => {
        const textDocument = state.textDocuments.get(uri)
        if (!textDocument) {
            throw new Error(`Text document does not exist with URI ${uri}`)
        }
        return textDocument
    }

    /**
     * Removes a model.
     *
     * @param uri The URI of the model to remove.
     */
    const removeTextDocument = (uri: string): void => {
        const model = getTextDocument(uri)
        state.textDocuments.delete(uri)
        if (state.languageReferences.decrement(model.languageId)) {
            state.activeLanguages.next(new Set<string>(state.languageReferences.keys()))
        }
    }

    /**
     * Returns the Viewer with the given viewerId.
     * Throws if no viewer exists with the given viewerId.
     */
    const getViewer = (viewerId: ViewerId['viewerId']): ExtensionViewer => {
        const viewer = state.viewComponents.get(viewerId)
        if (!viewer) {
            throw new ViewerNotFoundError(viewerId)
        }
        return viewer
    }

    const exposedToMain: FlatExtensionHostAPI = {
        haveInitialExtensionsLoaded: () => proxySubscribable(state.haveInitialExtensionsLoaded.asObservable()),

        // Configuration
        syncSettingsData: settings => state.settings.next(Object.freeze(settings)),

        // Workspace
        getWorkspaceRoots: () =>
            proxySubscribable(
                state.roots.pipe(
                    map(workspaceRoots =>
                        workspaceRoots.map(
                            ({ uri, inputRevision }): clientType.WorkspaceRoot => ({ uri: uri.href, inputRevision })
                        )
                    )
                )
            ),
        addWorkspaceRoot: root => {
            state.roots.next(Object.freeze([...state.roots.value, new ExtensionWorkspaceRoot(root)]))
            state.rootChanges.next()
        },
        removeWorkspaceRoot: uri => {
            state.roots.next(Object.freeze(state.roots.value.filter(workspace => workspace.uri.href !== uri)))
            state.rootChanges.next()
        },
        setSearchContext: context => {
            state.searchContext = context
            state.searchContextChanges.next(context)
        },

        // Search
        transformSearchQuery: query =>
            // TODO (simon) I don't enjoy the dark arts below
            // we return observable because of potential deferred addition of transformers
            // in this case we need to reissue the transformation and emit the resulting value
            // we probably won't need an Observable if we somehow coordinate with extensions activation
            proxySubscribable(
                state.queryTransformers.pipe(
                    switchMap(transformers =>
                        transformers.reduce(
                            (currentQuery: Observable<string>, transformer) =>
                                currentQuery.pipe(
                                    mergeMap(query => {
                                        const result = transformer.transformQuery(query)
                                        return result instanceof Promise ? from(result) : of(result)
                                    })
                                ),
                            of(query)
                        )
                    )
                )
            ),

        // Language
        getHover: (textParameters: TextDocumentPositionParameters) => {
            const document = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.hoverProviders,
                    providers => providersForDocument(document, providers, ({ selector }) => selector),
                    ({ provider }) => provider.provideHover(document, position),
                    results => fromHoverMerged(mergeProviderResults(results))
                )
            )
        },
        getDocumentHighlights: (textParameters: TextDocumentPositionParameters) => {
            const document = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.documentHighlightProviders,
                    providers => providersForDocument(document, providers, ({ selector }) => selector),
                    ({ provider }) => provider.provideDocumentHighlights(document, position),
                    mergeProviderResults
                ).pipe(map(result => (result.isLoading ? [] : result.result)))
            )
        },
        getDefinition: (textParameters: TextDocumentPositionParameters) => {
            const document = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.definitionProviders,
                    providers => providersForDocument(document, providers, ({ selector }) => selector),
                    ({ provider }) => provider.provideDefinition(document, position),
                    results => mergeProviderResults(results).map(fromLocation)
                )
            )
        },
        getReferences: (textParameters: TextDocumentPositionParameters, context: sourcegraph.ReferenceContext) => {
            const document = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.referenceProviders,
                    providers => providersForDocument(document, providers, ({ selector }) => selector),
                    ({ provider }) => provider.provideReferences(document, position, context),
                    results => mergeProviderResults(results).map(fromLocation)
                )
            )
        },
        hasReferenceProvidersForDocument: (textParameters: TextDocumentPositionParameters) => {
            const document = getTextDocument(textParameters.textDocument.uri)

            return proxySubscribable(
                state.referenceProviders.pipe(
                    map(providers => providersForDocument(document, providers, ({ selector }) => selector).length !== 0)
                )
            )
        },

        getLocations: (id: string, textParameters: TextDocumentPositionParameters) => {
            const document = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.locationProviders,
                    providers =>
                        providersForDocument(
                            document,
                            providers.filter(({ provider }) => id === provider.id),
                            ({ selector }) => selector
                        ),
                    ({ provider }) => provider.provider.provideLocations(document, position),
                    results => mergeProviderResults(results).map(fromLocation)
                )
            )
        },

        // Decorations
        getFileDecorations: (parameters: sourcegraph.FileDecorationContext) =>
            proxySubscribable(
                parameters.files.length === 0
                    ? EMPTY // Don't call providers when there are no files in the directory
                    : callProviders(
                          state.fileDecorationProviders,
                          identity,
                          // No need to filter
                          provider => provider.provideFileDecorations(parameters),
                          mergeProviderResults
                      ).pipe(
                          map(({ result }) =>
                              groupBy(
                                  result.filter(validateFileDecoration),
                                  // Get path from uri to key by path.
                                  // Path should always exist, but fall back to uri just in case
                                  ({ uri }) => parseRepoURI(uri).filePath || uri
                              )
                          )
                      )
            ),

        // MODELS

        //  TODO(tj): if not exists? doesn't seem that we can guarantee that just based on uri
        addViewerIfNotExists: viewerData => {
            const viewerId = `viewer#${state.lastViewerId++}`
            if (viewerData.type === 'CodeEditor') {
                state.modelReferences.increment(viewerData.resource)
            }
            let viewComponent: ExtensionViewer
            switch (viewerData.type) {
                case 'CodeEditor': {
                    const textDocument = getTextDocument(viewerData.resource)
                    viewComponent = new ExtensionCodeEditor({ ...viewerData, viewerId }, textDocument)
                    break
                }

                case 'DirectoryViewer': {
                    viewComponent = new ExtensionDirectoryViewer({ ...viewerData, viewerId })
                    break
                }
            }

            state.viewComponents.set(viewerId, viewComponent)
            if (viewerData.isActive) {
                state.activeViewComponentChanges.next(viewComponent)
            }
            state.viewerUpdates.next({ viewerId, action: 'addition', type: viewerData.type })
            return { viewerId }
        },

        removeViewer: ({ viewerId }) => {
            const viewer = getViewer(viewerId)
            state.viewComponents.delete(viewerId)
            // Check if this was the active viewer
            if (state.activeViewComponentChanges.value?.viewerId === viewerId) {
                state.activeViewComponentChanges.next(undefined)
            }
            state.viewerUpdates.next({ viewerId, action: 'removal', type: viewer.type })
            if (viewer.type === 'CodeEditor' && state.modelReferences.decrement(viewer.resource)) {
                removeTextDocument(viewer.resource)
            }
        },

        viewerUpdates: () => proxySubscribable(state.viewerUpdates.asObservable()),

        setEditorSelections: ({ viewerId }, selections) => {
            const viewer = getViewer(viewerId)
            assertViewerType(viewer, 'CodeEditor')
            viewer.update({ selections })
        },
        getTextDecorations: ({ viewerId }) => {
            const viewer = getViewer(viewerId)
            assertViewerType(viewer, 'CodeEditor')
            return proxySubscribable(viewer.mergedDecorations)
        },

        addTextDocumentIfNotExists: textDocumentData => {
            if (state.textDocuments.has(textDocumentData.uri)) {
                return
            }
            const textDocument = new ExtensionDocument(textDocumentData)
            state.textDocuments.set(textDocumentData.uri, textDocument)
            state.openedTextDocuments.next(textDocument)
            // Update activeLanguages if no other existing model has the same language.
            if (state.languageReferences.increment(textDocumentData.languageId)) {
                state.activeLanguages.next(new Set<string>(state.languageReferences.keys()))
            }
        },

        // For filtering visible panels by DocumentSelector
        getActiveViewComponentChanges: () => proxySubscribable(state.activeViewComponentChanges),

        // For panel view location provider arguments
        getActiveCodeEditorPosition: () =>
            proxySubscribable(
                state.activeViewComponentChanges.pipe(
                    switchMap(activeViewer => {
                        if (activeViewer?.type !== 'CodeEditor') {
                            return of(null)
                        }

                        return activeViewer.selectionsChanges.pipe(
                            map(selections => {
                                const sel = selections[0]
                                if (!sel) {
                                    return null
                                }

                                // TODO(sqs): Return null for empty selections (but currently all selected tokens are treated as an empty
                                // selection at the beginning of the token, so this would break a lot of things, so we only do this for empty
                                // selections when the start character is -1). HACK(sqs): Character === -1 means that the whole line is
                                // selected (this is a bug in the caller, but it is useful here).
                                const isEmpty =
                                    sel.start.line === sel.end.line &&
                                    sel.start.character === sel.end.character &&
                                    sel.start.character === -1
                                if (isEmpty) {
                                    return null
                                }

                                return {
                                    textDocument: { uri: activeViewer.resource },
                                    position: { line: sel.start.line, character: sel.start.character },
                                }
                            })
                        )
                    })
                )
            ),

        // Context data + Contributions
        updateContext: update => updateContext(update, state),
        registerContributions: rawContributions => {
            const parsedContributions = parseContributionExpressions(rawContributions)

            return proxy(addWithRollback(state.contributions, parsedContributions))
        },
        getContributions: ({ scope, extraContext, returnInactiveMenuItems }: ContributionOptions = {}) =>
            // TODO(tj): memoize access from mainthread (maybe by scope and extraContext (shallow))
            proxySubscribable(
                combineLatest([
                    state.contributions,
                    state.activeViewComponentChanges.pipe(
                        map((activeEditor): ViewerWithPartialModel | undefined => {
                            if (!activeEditor) {
                                return undefined
                            }

                            switch (activeEditor.type) {
                                case 'CodeEditor': {
                                    const { languageId } = getTextDocument(activeEditor.resource)
                                    return Object.assign(activeEditor, { model: { languageId } })
                                }
                                case 'DirectoryViewer':
                                    return activeEditor
                            }
                        })
                    ),
                    state.settings,
                    state.context as Subscribable<Context<unknown>>,
                ]).pipe(
                    map(([multiContributions, activeEditor, settings, context]) => {
                        // Merge in extra context.
                        if (extraContext) {
                            context = { ...context, ...extraContext }
                        }

                        // TODO(sqs): Observe context so that we update immediately upon changes.
                        const computedContext = computeContext(activeEditor, settings, context, scope)
                        return multiContributions.map(contributions => {
                            try {
                                const evaluatedContributions = evaluateContributions(computedContext, contributions)
                                return returnInactiveMenuItems
                                    ? evaluatedContributions
                                    : filterContributions(evaluatedContributions)
                            } catch (error) {
                                // An error during evaluation causes all of the contributions in the same entry to be
                                // discarded.
                                console.log('Discarding contributions: evaluating expressions or templates failed.', {
                                    contributions,
                                    error,
                                })
                                return {}
                            }
                        })
                    }),
                    map(mergeContributions),
                    distinctUntilChanged(isEqual)
                )
            ),

        // Notifications
        getPlainNotifications: () => proxySubscribable(state.plainNotifications.asObservable()),
        getProgressNotifications: () => proxySubscribable(state.progressNotifications.asObservable()),

        // Views
        getPanelViews: () =>
            // Don't need `combineLatestOrDefault` here since each panel view
            // is a BehaviorSubject, and therefore guaranteed to emit
            proxySubscribable(
                state.panelViews.pipe(
                    switchMap(panelViews => combineLatest([...panelViews])),
                    debounceTime(0)
                )
            ),

        // Insight page
        getInsightViewById: (id, context) =>
            proxySubscribable(
                state.insightsPageViewProviders.pipe(
                    switchMap(providers => {
                        const provider = providers.find(provider => {
                            // Get everything until last dot according to extension id naming convention
                            // <type>.<name>.<view type = directory|insightPage|homePage>
                            const providerId = provider.id.split('.').slice(0, -1).join('.')

                            return providerId === id
                        })

                        if (!provider) {
                            return throwError(new Error(`Couldn't find view with id ${id}`))
                        }

                        return providerResultToObservable(provider.viewProvider.provideView(context))
                    }),
                    catchError((error: unknown) => {
                        console.error('View provider errored:', error)
                        // Pass only primitive copied values because Error object is not
                        // cloneable in Firefox and Safari
                        const { message, name, stack } = asError(error)
                        return of({ message, name, stack } as ErrorLike)
                    }),
                    map(view => ({ id, view }))
                )
            ),

        getInsightsViews: (context, insightIds) =>
            getInsightsViews(context, state.insightsPageViewProviders, insightIds),

        getHomepageViews: context => proxySubscribable(callViewProviders(context, state.homepageViewProviders)),
        getDirectoryViews: context =>
            proxySubscribable(
                callViewProviders(
                    {
                        viewer: {
                            ...context.viewer,
                            directory: {
                                ...context.viewer.directory,
                                uri: new URL(context.viewer.directory.uri),
                            },
                        },
                        workspace: { uri: new URL(context.workspace.uri) },
                    },
                    state.directoryViewProviders
                )
            ),

        getGlobalPageViews: context => proxySubscribable(callViewProviders(context, state.globalPageViewProviders)),

        getStatusBarItems: ({ viewerId }) => {
            const viewer = getViewer(viewerId)
            if (viewer.type !== 'CodeEditor') {
                return proxySubscribable(EMPTY)
            }

            return proxySubscribable(
                viewer.mergedStatusBarItems.pipe(
                    debounceTime(0),
                    map(statusBarItems =>
                        statusBarItems.sort(
                            (a, b) => a.text[0].toLowerCase().charCodeAt(0) - b.text[0].toLowerCase().charCodeAt(0)
                        )
                    )
                )
            )
        },

        // Content
        getLinkPreviews: (url: string) =>
            proxySubscribable(
                callProviders(
                    state.linkPreviewProviders,
                    entries => entries.filter(entry => url.startsWith(entry.urlMatchPattern)),
                    ({ provider }) => provider.provideLinkPreview(new URL(url)),
                    stuffs => mergeLinkPreviews(stuffs)
                ).pipe(map(result => (result.isLoading ? null : result.result)))
            ),

        getActiveExtensions: () => proxySubscribable(state.activeExtensions),
    }

    return exposedToMain
}

export interface RegisteredProvider<T> {
    selector: sourcegraph.DocumentSelector
    provider: T
}

// TODO (loic, felix) it might make sense to port tests with the rest of provider registries.
/**
 * Filters a list of Providers (P type) based on their selectors and a document
 *
 * @param document to use for filtering
 * @param entries array of providers (P[])
 * @param selector a way to get a selector from a Provider
 * @returns a filtered array of providers
 */
export function providersForDocument<P>(
    document: TextDocumentIdentifier,
    entries: readonly P[],
    selector: (p: P) => sourcegraph.DocumentSelector
): P[] {
    return entries.filter(provider =>
        match(selector(provider), {
            uri: document.uri,
            languageId: getModeFromPath(parseRepoURI(document.uri).filePath || ''),
        })
    )
}

/**
 * Helper function to abstract common logic of invoking language providers.
 *
 * 1. filters providers
 * 2. invokes filtered providers via invokeProvider function
 * 3. adds [LOADING] state for each provider result stream
 * 4. omits errors from provider results with potential logging
 * 5. aggregates latests results from providers based on mergeResult function
 *
 * @param providersObservable observable of provider collection (expected to emit if a provider was added or removed)
 * @param filterProviders specifies which providers should be invoked
 * @param invokeProvider specifies how to get results from a provider (usually a closure over provider arguments)
 * @param mergeResult specifies how providers results should be aggregated
 * @param logErrors if console.error should be used for reporting errors from providers
 * @returns observable of aggregated results from all providers based on mergeProviderResults function
 */
export function callProviders<TRegisteredProvider, TProviderResult, TMergedResult>(
    providersObservable: Observable<readonly TRegisteredProvider[]>,
    filterProviders: (providers: readonly TRegisteredProvider[]) => TRegisteredProvider[],
    invokeProvider: (provider: TRegisteredProvider) => sourcegraph.ProviderResult<TProviderResult>,
    mergeResult: (providerResults: readonly (TProviderResult | 'loading' | null | undefined)[]) => TMergedResult,
    logErrors: boolean = true
): Observable<MaybeLoadingResult<TMergedResult>> {
    const logError = (...args: any): void => {
        if (logErrors) {
            console.error('Provider errored:', ...args)
        }
    }
    const safeInvokeProvider = (provider: TRegisteredProvider): sourcegraph.ProviderResult<TProviderResult> => {
        try {
            return invokeProvider(provider)
        } catch (error) {
            logError(error)
            return null
        }
    }

    return providersObservable
        .pipe(
            map(providers => filterProviders(providers)),

            switchMap(providers =>
                combineLatestOrDefault(
                    providers.map(provider =>
                        concat(
                            [LOADING],
                            providerResultToObservable(safeInvokeProvider(provider)).pipe(
                                defaultIfEmpty<typeof LOADING | TProviderResult | null | undefined>(null),
                                catchError(error => {
                                    logError(error)
                                    return [null]
                                })
                            )
                        )
                    )
                )
            )
        )
        .pipe(
            defaultIfEmpty<(typeof LOADING | TProviderResult | null | undefined)[]>([]),
            map(results => ({
                isLoading: results.some(hover => hover === LOADING),
                result: mergeResult(results),
            })),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )
}

/**
 * Merges provider results
 *
 * @param results latest results from providers
 * @template TProviderResultElement Type of an element of the provider result array
 */
export function mergeProviderResults<TProviderResultElement>(
    results: readonly (typeof LOADING | TProviderResultElement | TProviderResultElement[] | null | undefined)[]
): TProviderResultElement[] {
    return results
        .filter(isNot(isExactly(LOADING)))
        .flatMap(castArray)
        .filter(isDefined)
}

/** Object of array of file decorations keyed by path relative to repo root uri */
export type FileDecorationsByPath = Record<string, sourcegraph.FileDecoration[] | undefined>

// Viewers + documents

const VIEWER_NOT_FOUND_ERROR_NAME = 'ViewerNotFoundError'
class ViewerNotFoundError extends Error {
    public readonly name = VIEWER_NOT_FOUND_ERROR_NAME
    constructor(viewerId: string) {
        super(`Viewer not found: ${viewerId}`)
    }
}

function assertViewerType<T extends ExtensionViewer['type']>(
    viewer: ExtensionViewer,
    type: T
): asserts viewer is ExtensionViewer & { type: T } {
    if (viewer.type !== type) {
        throw new Error(`Viewer ID ${viewer.viewerId} is type ${viewer.type}, expected ${type}`)
    }
}

// Content

interface MarkupContentPlainTextOnly extends Pick<sourcegraph.MarkupContent, 'value'> {
    kind: sourcegraph.MarkupKind.PlainText
}

/**
 * Represents one or more {@link sourcegraph.LinkPreview} values merged together.
 */
export interface LinkPreviewMerged {
    /** The content of the merged {@link sourcegraph.LinkPreview} values. */
    content: sourcegraph.MarkupContent[]

    /** The hover content of the merged {@link sourcegraph.LinkPreview} values. */
    hover: MarkupContentPlainTextOnly[]
}

function mergeLinkPreviews(
    values: readonly (sourcegraph.LinkPreview | 'loading' | null | undefined)[]
): LinkPreviewMerged | null {
    const nonemptyValues = values
        .filter(isDefined)
        .filter((value): value is Exclude<sourcegraph.LinkPreview | 'loading', 'loading'> => value !== 'loading')
    const contentValues = nonemptyValues.filter(property('content', isDefined))
    const hoverValues = nonemptyValues.filter(property('hover', isDefined))
    if (hoverValues.length === 0 && contentValues.length === 0) {
        return null
    }
    return { content: contentValues.map(({ content }) => content), hover: hoverValues.map(({ hover }) => hover) }
}

// Views

/**
 * A map from type of container names to the internal type of the context parameter provided by the container.
 */
export interface ViewContexts {
    [ContributableViewContainer.Panel]: never
    [ContributableViewContainer.Homepage]: {}
    [ContributableViewContainer.InsightsPage]: {}
    [ContributableViewContainer.GlobalPage]: Record<string, string>
    [ContributableViewContainer.Directory]: sourcegraph.DirectoryViewContext
}

export interface RegisteredViewProvider<W extends ContributableViewContainer> {
    id: string
    viewProvider: {
        provideView: (context: ViewContexts[W]) => sourcegraph.ProviderResult<sourcegraph.View>
    }
}

function callViewProviders<W extends ContributableViewContainer>(
    context: ViewContexts[W],
    providers: Observable<readonly RegisteredViewProvider<W>[]>
): Observable<ViewProviderResult[]> {
    return providers.pipe(
        debounceTime(0),
        switchMap(providers =>
            combineLatest([
                of(null),
                ...providers.map(({ viewProvider, id }) =>
                    concat(
                        [undefined],
                        providerResultToObservable(viewProvider.provideView(context)).pipe(
                            defaultIfEmpty<sourcegraph.View | null | undefined>(null),
                            catchError((error: unknown): [ErrorLike] => {
                                console.error('View provider errored:', error)
                                // Pass only primitive copied values because Error object is not
                                // cloneable in Firefox and Safari
                                const { message, name, stack } = asError(error)
                                return [{ message, name, stack } as ErrorLike]
                            })
                        )
                    ).pipe(map(view => ({ id, view })))
                ),
            ])
        ),
        map(views => views.filter(allOf(isDefined, property('view', isNot(isExactly(null))))))
    )
}

/**
 * A workspace root with additional metadata that is not exposed to extensions.
 */
export interface WorkspaceRootWithMetadata extends clientType.WorkspaceRoot {
    /**
     * The original input Git revision that the user requested. The {@link WorkspaceRoot#uri} value will contain
     * the Git commit SHA resolved from the input revision, but it is useful to also know the original revision
     * (e.g., to construct URLs for the user that don't result in them navigating from a branch view to a commit
     * SHA view).
     *
     * For example, if the user is viewing the web page https://github.com/alice/myrepo/blob/master/foo.js (note
     * that the URL contains a Git revision "master"), the input revision is "master".
     *
     * The empty string is a valid value (meaning that the default should be used, such as "HEAD" in Git) and is
     * distinct from undefined. If undefined, the Git commit SHA from {@link WorkspaceRoot#uri} should be used.
     */
    inputRevision?: string
}

/** @internal */
export interface PanelViewData extends Omit<sourcegraph.PanelView, 'unsubscribe'> {
    id: string
}

/**
 * A notification message to display to the user.
 */
export type ExtensionNotification = PlainNotification | ProgressNotification

interface BaseNotification {
    /** The message of the notification. */
    message?: string

    /**
     * The type of the message.
     */
    type: sourcegraph.NotificationType

    /** The source of the notification.  */
    source?: string
}

export interface PlainNotification extends BaseNotification {}

export interface ProgressNotification {
    // Put all base notification properties in a nested object because
    // ProgressNotifications are proxied, so it's better to clone this
    // notification object than to wait for all property access promises
    // to resolve
    baseNotification: BaseNotification

    /**
     * Progress updates to show in this notification (progress bar and status messages).
     * If this Observable errors, the notification will be changed to an error type.
     */
    progress: ProxySubscribable<sourcegraph.Progress>
}

export interface ViewProviderResult {
    /** The ID of the view provider. */
    id: string

    /** The result returned by the provider. */
    view: sourcegraph.View | undefined | ErrorLike
}

/**
 * The type of a notification.
 * This is needed because if sourcegraph.NotificationType enum values are referenced,
 * the `sourcegraph` module import at the top of the file is emitted in the generated code.
 */
export const NotificationType: typeof sourcegraph.NotificationType = {
    Error: 1,
    Warning: 2,
    Info: 3,
    Log: 4,
    Success: 5,
}

// Contributions

export interface ContributionOptions<T = unknown> {
    scope?: ContributionScope | undefined
    extraContext?: Context<T>
    returnInactiveMenuItems?: boolean
}
