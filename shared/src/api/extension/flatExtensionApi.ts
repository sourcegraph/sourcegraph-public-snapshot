import { SettingsCascade } from '../../settings/settings'
import { Remote, proxy } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { BehaviorSubject, Subject, of, Observable, from, concat } from 'rxjs'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { syncSubscription } from '../util'
import { switchMap, mergeMap, map, defaultIfEmpty, catchError, distinctUntilChanged } from 'rxjs/operators'
import { proxySubscribable, providerResultToObservable } from './api/common'
import { TextDocumentIdentifier, match } from '../client/types/textDocument'
import { getModeFromPath } from '../../languages'
import { parseRepoURI } from '../../util/url'
import { toPosition } from './api/types'
import { TextDocumentPositionParams } from '../protocol'
import { LOADING, MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { combineLatestOrDefault } from '../../util/rxjs/combineLatestOrDefault'
import { Hover } from '@sourcegraph/extension-api-types'
import { isEqual } from 'lodash'
import { fromHoverMerged, HoverMerged } from '../client/types/hover'
import { isNot, isExactly } from '../../util/types'
import { ViewerUpdate, ViewerId, ExtensionViewer } from '../viewerTypes'
import { ReferenceCounter } from '../../util/ReferenceCounter'
import { ExtensionDocument } from './api/textDocument'
import { ExtensionCodeEditor } from './api/codeEditor'
import { ClientCodeEditorAPI } from '../client/api/codeEditor'
import { ExtensionDirectoryViewer } from './api/directoryViewer'

/**
 * Holds the entire state exposed to the extension host
 * as a single object
 */
export interface ExtState {
    settings: Readonly<SettingsCascade<object>>

    // Workspace
    roots: readonly sourcegraph.WorkspaceRoot[]
    versionContext: string | undefined

    // Search
    queryTransformers: BehaviorSubject<sourcegraph.QueryTransformer[]>

    // Lang
    hoverProviders: BehaviorSubject<RegisteredProvider<sourcegraph.HoverProvider>[]>
}

export interface RegisteredProvider<T> {
    selector: sourcegraph.DocumentSelector
    provider: T
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: typeof sourcegraph.workspace
    exposedToMain: FlatExtHostAPI
    // todo this is needed as a temp solution for getter problem
    state: Readonly<ExtState>
    commands: typeof sourcegraph['commands']
    search: typeof sourcegraph['search']
    languages: Pick<typeof sourcegraph['languages'], 'registerHoverProvider'>
}

/**
 * mimics sourcegraph.workspace namespace without documents
 */
export type PartialWorkspaceNamespace = Omit<typeof sourcegraph['workspace'], 'roots' | 'versionContext'>

const VIEWER_NOT_FOUND_ERROR_NAME = 'ViewerNotFoundError'
class ViewerNotFoundError extends Error {
    public readonly name = VIEWER_NOT_FOUND_ERROR_NAME
    constructor(viewerId: string) {
        super(`Viewer not found: ${viewerId}`)
    }
}

/**
 * Holds internally ExtState and manages communication with the Client
 * Returns the initialized public extension API pieces ready for consumption and the internal extension host API ready to be exposed to the main thread
 * NOTE that this function will slowly merge with the one in extensionHost.ts
 *
 * @param mainAPI
 */
export const initNewExtensionAPI = (
    mainAPI: Remote<MainThreadAPI>,
    initialSettings: Readonly<SettingsCascade<object>>,
    codeEditorAPI: Remote<ClientCodeEditorAPI>
): InitResult => {
    const state: ExtState = {
        roots: [],
        versionContext: undefined,
        settings: initialSettings,
        queryTransformers: new BehaviorSubject<sourcegraph.QueryTransformer[]>([]),
        hoverProviders: new BehaviorSubject<RegisteredProvider<sourcegraph.HoverProvider>[]>([]),
    }

    const configChanges = new BehaviorSubject<void>(undefined)
    // Most extensions never call `configuration.get()` synchronously in `activate()` to get
    // the initial settings data, and instead only subscribe to configuration changes.
    // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.

    const rootChanges = new Subject<void>()

    const versionContextChanges = new Subject<string | undefined>()

    let lastViewerId = 0

    /** Mutable map of viewer ID to viewer. */
    const viewComponents = new Map<string, ExtensionCodeEditor | ExtensionDirectoryViewer>()
    const activeViewComponentChanges = new BehaviorSubject<ExtensionViewer | undefined>(undefined)

    /** A map of URIs to text documents */
    const textDocuments = new Map<string, ExtensionDocument>()
    const openedTextDocuments = new Subject<ExtensionDocument>()
    const activeLanguages = new BehaviorSubject<ReadonlySet<string>>(new Set())
    const languageReferences = new ReferenceCounter<string>()

    const getTextDocument = (uri: string): ExtensionDocument => {
        const textDocument = textDocuments.get(uri)
        if (!textDocument) {
            throw new Error(`Text document does not exist with URI ${uri}`)
        }
        return textDocument
    }

    /**
     * Returns the Viewer with the given viewerId.
     * Throws if no viewer exists with the given viewerId.
     */
    const getViewer = (viewerId: ViewerId['viewerId']): ExtensionViewer => {
        const viewer = viewComponents.get(viewerId)
        if (!viewer) {
            throw new ViewerNotFoundError(viewerId)
        }
        return viewer
    }

    /**
     * Removes a model.
     *
     * @param uri The URI of the model to remove.
     */
    const removeTextDocument = (uri: string): void => {
        const model = getTextDocument(uri)
        textDocuments.delete(uri)
        if (languageReferences.decrement(model.languageId)) {
            activeLanguages.next(new Set<string>(languageReferences.keys()))
        }
    }

    const modelReferences = new ReferenceCounter()

    const exposedToMain: FlatExtHostAPI = {
        // Configuration
        syncSettingsData: data => {
            state.settings = Object.freeze(data)
            configChanges.next()
        },

        // Workspace
        syncRoots: (roots): void => {
            state.roots = Object.freeze(roots.map(plain => ({ ...plain, uri: new URL(plain.uri) })))
            rootChanges.next()
        },
        syncVersionContext: context => {
            state.versionContext = context
            versionContextChanges.next(context)
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
        getHover: (textParameters: TextDocumentPositionParams) => {
            const textDocument = getTextDocument(textParameters.textDocument.uri)
            const position = toPosition(textParameters.position)

            return proxySubscribable(
                callProviders(
                    state.hoverProviders,
                    textDocument,
                    provider => provider.provideHover(textDocument, position),
                    mergeHoverResults
                )
            )
        },

        // Viewer
        addViewerIfNotExists: viewerData => {
            const viewerId = `viewer#${++lastViewerId}`
            if (viewerData.type === 'CodeEditor') {
                modelReferences.increment(viewerData.resource)
            }
            let viewComponent: ExtensionViewer
            switch (viewerData.type) {
                case 'CodeEditor': {
                    const textDocument = getTextDocument(viewerData.resource)
                    viewComponent = new ExtensionCodeEditor({ ...viewerData, viewerId }, codeEditorAPI, textDocument)
                    break
                }
                case 'DirectoryViewer': {
                    viewComponent = new ExtensionDirectoryViewer({ ...viewerData, viewerId })
                    break
                }
            }
            viewComponents.set(viewerId, viewComponent)
            if (viewerData.isActive) {
                activeViewComponentChanges.next(viewComponent)
            }
            if (viewerData.isActive) {
                activeViewComponentChanges.next(viewComponent)
            }
            return { viewerId }
        },
        removeViewer: ({ viewerId }) => {
            const viewer = getViewer(viewerId)
            viewComponents.delete(viewerId)
            // Check if this was the active viewer
            if (activeViewComponentChanges.value?.viewerId === viewerId) {
                activeViewComponentChanges.next(undefined)
            }
            if (viewer.type === 'CodeEditor' && modelReferences.decrement(viewer.resource)) {
                removeTextDocument(viewer.resource)
            }
        },
        getActiveCodeEditorPosition: () =>
            proxySubscribable(
                activeViewComponentChanges.pipe(
                    map(activeViewer => {
                        if (activeViewer?.type !== 'CodeEditor') {
                            return null
                        }
                        const sel = activeViewer.selections[0]
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
                            position: sel.start,
                        }
                    })
                )
            ),
        setEditorSelections: ({ viewerId }, selections) => {
            const viewer = getViewer(viewerId)
            if (viewer.type !== 'CodeEditor') {
                throw new Error(`Editor ID ${viewerId} is type ${String(viewer.type)}, expected CodeEditor`)
            }
            viewer.update({ selections })
        },

        // Text documents
        addTextDocumentIfNotExists: textDocumentData => {
            if (textDocuments.has(textDocumentData.uri)) {
                return
            }
            const textDocument = new ExtensionDocument(textDocumentData)
            textDocuments.set(textDocumentData.uri, textDocument)
            openedTextDocuments.next(textDocument)
            // Update activeLanguages if no other existing model has the same language.
            if (languageReferences.increment(textDocumentData.languageId)) {
                activeLanguages.next(new Set<string>(languageReferences.keys()))
            }
        },
    }

    // Configuration
    const getConfiguration = <C extends object>(): sourcegraph.Configuration<C> => {
        const snapshot = state.settings.final as Readonly<C>

        const configuration: sourcegraph.Configuration<C> & { toJSON: any } = {
            value: snapshot,
            get: key => snapshot[key],
            update: (key, value) => mainAPI.applySettingsEdit({ path: [key as string | number], value }),
            toJSON: () => snapshot,
        }
        return configuration
    }

    // Workspace
    const workspace: typeof sourcegraph.workspace = {
        get roots() {
            return state.roots
        },
        get versionContext() {
            return state.versionContext
        },
        get textDocuments() {
            return [...textDocuments.values()]
        },
        openedTextDocuments,
        onDidOpenTextDocument: openedTextDocuments,
        onDidChangeRoots: rootChanges.asObservable(),
        rootChanges: rootChanges.asObservable(),
        versionContextChanges: versionContextChanges.asObservable(),
    }

    // Commands
    const commands: typeof sourcegraph['commands'] = {
        executeCommand: (command, ...args) => mainAPI.executeCommand(command, args),
        registerCommand: (command, callback) => syncSubscription(mainAPI.registerCommand(command, proxy(callback))),
    }

    // Search
    const search: typeof sourcegraph['search'] = {
        registerQueryTransformer: transformer => addWithRollback(state.queryTransformers, transformer),
    }

    // Languages
    const registerHoverProvider = (
        selector: sourcegraph.DocumentSelector,
        provider: sourcegraph.HoverProvider
    ): sourcegraph.Unsubscribable => addWithRollback(state.hoverProviders, { selector, provider })

    return {
        exposedToMain,
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        workspace,
        state,
        commands,
        search,
        languages: {
            registerHoverProvider,
        },
    }
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
function providersForDocument<P>(
    document: TextDocumentIdentifier,
    entries: P[],
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
 * calls next() on behaviorSubject with a immutably added element ([...old, value])
 *
 * @param behaviorSubject subject that holds a collection
 * @param value to add to a collection
 * @returns Unsubscribable that will remove that element from the behaviorSubject.value and call next() again
 */
function addWithRollback<T>(behaviorSubject: BehaviorSubject<T[]>, value: T): sourcegraph.Unsubscribable {
    behaviorSubject.next([...behaviorSubject.value, value])
    return {
        unsubscribe: () => behaviorSubject.next(behaviorSubject.value.filter(item => item !== value)),
    }
}

/**
 * Helper function to abstract common logic of invoking language providers.
 *
 * 1. filters providers based on document
 * 2. invokes filtered providers via invokeProvider function
 * 3. adds [LOADING] state for each provider result stream
 * 4. omits errors from provider results with potential logging
 * 5. aggregates latests results from providers based on mergeResult function
 *
 * @param providersObservable observable of provider collection (expected to emit if a provider was added or removed)
 * @param document used for filtering providers
 * @param invokeProvider specifies how to get results from a provider (usually a closure over provider arguments)
 * @param mergeResult specifies how providers results should be aggregated
 * @param logErrors if console.error should be used for reporting errors from providers
 * @returns observable of aggregated results from all providers based on mergeResults function
 */
export function callProviders<TProvider, TProviderResult, TMergedResult>(
    providersObservable: Observable<RegisteredProvider<TProvider>[]>,
    document: TextDocumentIdentifier,
    invokeProvider: (provider: TProvider) => sourcegraph.ProviderResult<TProviderResult>,
    mergeResult: (providerResults: (TProviderResult | 'loading' | null | undefined)[]) => TMergedResult,
    logErrors: boolean = true
): Observable<MaybeLoadingResult<TMergedResult>> {
    return providersObservable
        .pipe(
            map(providers => providersForDocument(document, providers, ({ selector }) => selector)),
            switchMap(providers =>
                combineLatestOrDefault(
                    providers.map(provider =>
                        concat(
                            [LOADING],
                            providerResultToObservable(invokeProvider(provider.provider)).pipe(
                                defaultIfEmpty<typeof LOADING | TProviderResult | null | undefined>(null),
                                catchError(error => {
                                    if (logErrors) {
                                        console.error('Provider errored:', error)
                                    }
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
 * merges latests results from hover providers into a form that is convenient to show
 *
 * @param results latests results from hover providers
 * @returns a {@link HoverMerged} results if there are any actual Hover results or null in case of no results or loading
 */
export function mergeHoverResults(results: (typeof LOADING | Hover | null | undefined)[]): HoverMerged | null {
    return fromHoverMerged(results.filter(isNot(isExactly(LOADING))))
}
