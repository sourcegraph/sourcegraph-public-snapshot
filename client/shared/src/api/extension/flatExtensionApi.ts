import { SettingsCascade } from '../../settings/settings'
import { Remote, proxy, ProxyMarked } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import {
    BehaviorSubject,
    Subject,
    of,
    Observable,
    from,
    concat,
    EMPTY,
    ReplaySubject,
    combineLatest,
    Subscribable,
    throwError,
} from 'rxjs'
import {
    FlatExtensionHostAPI,
    MainThreadAPI,
    NotificationType,
    PlainNotification,
    ProgressNotification,
} from '../contract'
import { syncSubscription } from '../util'
import {
    switchMap,
    mergeMap,
    map,
    defaultIfEmpty,
    catchError,
    distinctUntilChanged,
    mapTo,
    tap,
    shareReplay,
    withLatestFrom,
    bufferTime,
    throttleTime,
    debounceTime,
} from 'rxjs/operators'
import { proxySubscribable, providerResultToObservable } from './api/common'
import { TextDocumentIdentifier, match } from '../client/types/textDocument'
import { getModeFromPath } from '../../languages'
import { parseRepoURI } from '../../util/url'
import { ExtensionDocuments } from './api/documents'
import { fromLocation, toPosition } from './api/types'
import { Contributions, TextDocumentPositionParameters } from '../protocol'
import { LOADING, MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { combineLatestOrDefault } from '../../util/rxjs/combineLatestOrDefault'
import { castArray, groupBy, identity, isEqual, isMatch, sortBy } from 'lodash'
import { fromHoverMerged } from '../client/types/hover'
import { isNot, isExactly, isDefined } from '../../util/types'
import { validateFileDecoration } from './api/decorations'
import { InitData } from './extensionHost'
import { ExtensionDocument } from './api/textDocument'
import { ReferenceCounter } from '../../util/ReferenceCounter'
import { ExtensionCodeEditor } from './api/codeEditor'
import { ExtensionViewer, ViewerId } from '../viewerTypes'
import { ExtensionDirectoryViewer } from './api/directoryViewer'
import { ExtensionWorkspaceRoot } from './api/workspaceRoot'
import { asError, createAggregateError, isErrorLike } from '../../util/errors'
import { ViewerWithPartialModel } from '../client/services/viewerService'
import { computeContext } from '../client/context/context'
import {
    filterContributions,
    evaluateContributions,
    mergeContributions,
    parseContributionExpressions,
} from '../client/services/contribution'
import { wrapRemoteObservable } from '../client/api/common'
import {
    ConfiguredRegistryExtension,
    toConfiguredRegistryExtension,
    ConfiguredExtension,
    isExtensionEnabled,
    getScriptURLFromExtensionManifest,
} from '../../extensions/extension'
import { gql } from '../../graphql/graphql'
import * as GQL from '../../graphql/schema'
import { ExtensionManifest } from '../../extensions/extensionManifest'
import { fromFetch } from 'rxjs/fetch'
import { checkOk } from '../../backend/fetch'
import { memoizeObservable } from '../../util/memoizeObservable'
import { ExecutableExtension } from '../client/services/extensionsService'

/**
 * Holds the entire state exposed to the extension host
 * as a single object
 */
export interface ExtensionHostState {
    settings: BehaviorSubject<Readonly<SettingsCascade>>

    // Workspace
    roots: BehaviorSubject<readonly ExtensionWorkspaceRoot[]>
    versionContext: BehaviorSubject<string | undefined>

    // Search
    queryTransformers: BehaviorSubject<readonly sourcegraph.QueryTransformer[]>

    // Lang
    hoverProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.HoverProvider>[]>
    documentHighlightProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.DocumentHighlightProvider>[]>
    definitionProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.DefinitionProvider>[]>

    // Decorations
    fileDecorationProviders: BehaviorSubject<readonly sourcegraph.FileDecorationProvider[]>

    // Context + Contributions
    context: BehaviorSubject<Context>
    /** All contributions, including those that are not enabled in the current context. */
    contributions: BehaviorSubject<readonly Contributions[]>

    // Viewer + Text documents
    lastViewerId: number
    openedTextDocuments: Subject<ExtensionDocument>
    activeLanguages: BehaviorSubject<ReadonlySet<string>>
    /** TODO(tj): URI? */
    modelReferences: ReferenceCounter<string>
    languageReferences: ReferenceCounter<string>
    /** Mutable map of URIs to text documents */
    textDocuments: Map<string, ExtensionDocument>

    /** Mutable map of viewer ID to viewer. */
    viewComponents: Map<string, ExtensionViewer> // TODO(tj): ext dir viewer
    activeViewComponentChanges: BehaviorSubject<ExtensionViewer | undefined>

    plainNotifications: ReplaySubject<PlainNotification>
    progressNotifications: ReplaySubject<ProgressNotification & ProxyMarked>
}

export interface RegisteredProvider<T> {
    selector: sourcegraph.DocumentSelector
    provider: T
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: typeof sourcegraph['workspace']
    exposedToMain: FlatExtensionHostAPI
    // todo this is needed as a temp solution for getter problem
    state: Readonly<ExtensionHostState>
    commands: typeof sourcegraph['commands']
    search: typeof sourcegraph['search']
    languages: Pick<
        typeof sourcegraph['languages'],
        'registerHoverProvider' | 'registerDocumentHighlightProvider' | 'registerDefinitionProvider'
    >
    graphQL: typeof sourcegraph['graphQL']
    internal: Pick<typeof sourcegraph['internal'], 'updateContext'>
    app: Omit<typeof sourcegraph['app'], 'createDecorationType' | 'createPanelView' | 'registerViewProvider'>
}

/** Object of array of file decorations keyed by path relative to repo root uri */
export type FileDecorationsByPath = Record<string, sourcegraph.FileDecoration[] | undefined>

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

/**
 * Context is an arbitrary, immutable set of key-value pairs. Its value can be any JSON object.
 *
 * @template T If you have a value with a property of type T that is not one of the primitive types listed below
 * (or Context), you can use Context<T> to hold that value. T must be a value that can be represented by a JSON
 * object.
 */
export interface Context<T = never>
    extends Record<
        string,
        string | number | boolean | null | Context<T> | T | (string | number | boolean | null | Context<T> | T)[]
    > {}

/** A registered set of contributions from an extension in the registry. */
export interface ContributionsEntry {
    /**
     * The contributions, either as a value or an observable.
     *
     * If an observable is used, it should be a cold Observable and emit (e.g., its current value) upon
     * subscription. The {@link ContributionRegistry#contributions} observable blocks until all observables have
     * emitted.
     */
    contributions: Contributions | Observable<Contributions | Contributions[]>
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
    { initialSettings, clientApplication }: Pick<InitData, 'initialSettings' | 'clientApplication'>,
    textDocuments: ExtensionDocuments
): InitResult => {
    const state: ExtensionHostState = {
        roots: new BehaviorSubject<readonly ExtensionWorkspaceRoot[]>([]),
        versionContext: new BehaviorSubject<string | undefined>(undefined),

        // Most extensions never call `configuration.get()` synchronously in `activate()` to get
        // the initial settings data, and instead only subscribe to configuration changes.
        // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.
        settings: new BehaviorSubject<Readonly<SettingsCascade<object>>>(initialSettings),

        queryTransformers: new BehaviorSubject<readonly sourcegraph.QueryTransformer[]>([]),

        hoverProviders: new BehaviorSubject<readonly RegisteredProvider<sourcegraph.HoverProvider>[]>([]),
        documentHighlightProviders: new BehaviorSubject<
            readonly RegisteredProvider<sourcegraph.DocumentHighlightProvider>[]
        >([]),
        definitionProviders: new BehaviorSubject<readonly RegisteredProvider<sourcegraph.DefinitionProvider>[]>([]),
        fileDecorationProviders: new BehaviorSubject<readonly sourcegraph.FileDecorationProvider[]>([]),

        context: new BehaviorSubject<Context>({
            'clientApplication.isSourcegraph': clientApplication === 'sourcegraph',

            // Arbitrary, undocumented versioning for extensions that need different behavior for different
            // Sourcegraph versions.
            //
            // TODO: Make this more advanced if many extensions need this (although we should try to avoid
            // extensions needing this).
            'clientApplication.extensionAPIVersion.major': 3,
        }),
        contributions: new BehaviorSubject<readonly Contributions[]>([]),

        lastViewerId: 0,
        textDocuments: new Map<string, ExtensionDocument>(),
        openedTextDocuments: new Subject<ExtensionDocument>(),
        viewComponents: new Map<string, ExtensionCodeEditor>(),

        activeLanguages: new BehaviorSubject<ReadonlySet<string>>(new Set()),
        languageReferences: new ReferenceCounter<string>(),
        modelReferences: new ReferenceCounter<string>(),

        activeViewComponentChanges: new BehaviorSubject<ExtensionViewer | undefined>(undefined),

        // Use ReplaySubject so we don't lose notifications in case the client application subscribes
        // to notification streams after extensions have already sent notifications.
        // There should be no issue re: stale notifications, since client applications should only
        // create one "notification manager" instance.
        plainNotifications: new ReplaySubject<PlainNotification>(3),
        progressNotifications: new ReplaySubject<ProgressNotification & ProxyMarked>(3),
    }

    const getTextDocument = (uri: string): ExtensionDocument => {
        const textDocument = textDocuments.get(uri)
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
        // Configuration
        syncSettingsData: settings => state.settings.next(Object.freeze(settings)),

        // Workspace
        getWorkspaceRoots: () => state.roots.value.map(({ uri, inputRevision }) => ({ uri: uri.href, inputRevision })),
        addWorkspaceRoot: root =>
            state.roots.next(Object.freeze([...state.roots.value, new ExtensionWorkspaceRoot(root)])),
        removeWorkspaceRoot: uri =>
            state.roots.next(Object.freeze(state.roots.value.filter(workspace => workspace.uri.href !== uri))),
        setVersionContext: context => state.versionContext.next(context),

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
            const document = textDocuments.get(textParameters.textDocument.uri)
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
            const document = textDocuments.get(textParameters.textDocument.uri)
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
            const document = textDocuments.get(textParameters.textDocument.uri)
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

        //  TODO(tj): if not exists?
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
            return { viewerId }
        },

        removeViewer: ({ viewerId }) => {
            const viewer = getViewer(viewerId)
            state.viewComponents.delete(viewerId)
            // Check if this was the active viewer
            if (state.activeViewComponentChanges.value?.viewerId === viewerId) {
                state.activeViewComponentChanges.next(undefined)
            }
            if (viewer.type === 'CodeEditor' && state.modelReferences.decrement(viewer.resource)) {
                removeTextDocument(viewer.resource)
            }
        },

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

        // TODO(tj): for panel view location provider arguments
        getActiveCodeEditorPosition: () =>
            proxySubscribable(
                state.activeViewComponentChanges.pipe(
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

        // Context data + Contributions
        updateContext,
        registerContributions: rawContributions => {
            const parsedContributions = parseContributionExpressions(rawContributions)

            return proxy(addWithRollback(state.contributions, parsedContributions))
        },
        getContributions: (scope, extraContext) =>
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
                                console.log({ computedContext, contributions })
                                return filterContributions(evaluateContributions(computedContext, contributions))
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
    }

    // Configuration
    const getConfiguration = <C extends object>(): sourcegraph.Configuration<C> => {
        const snapshot = state.settings.value.final as Readonly<C>

        const configuration: sourcegraph.Configuration<C> & { toJSON: any } = {
            value: snapshot,
            get: key => snapshot[key],
            update: (key, value) => mainAPI.applySettingsEdit({ path: [key as string | number], value }),
            toJSON: () => snapshot,
        }
        return configuration
    }

    // Workspace
    const workspace: typeof sourcegraph['workspace'] = {
        get textDocuments() {
            return [...state.textDocuments.values()]
        },
        get roots() {
            return state.roots.value
        },
        get versionContext() {
            return state.versionContext.value
        },
        onDidOpenTextDocument: state.openedTextDocuments.asObservable(),
        openedTextDocuments: state.openedTextDocuments.asObservable(),
        onDidChangeRoots: state.roots.pipe(mapTo(undefined)),
        rootChanges: state.roots.pipe(mapTo(undefined)),
        versionContextChanges: state.versionContext.pipe(mapTo(undefined)),
    }

    const createProgressReporter = async (
        options: sourcegraph.ProgressOptions
        // `showProgress` returned a promise when progress reporters were created
        // in the main thread. continue to return promise for backward compatibility
        // eslint-disable-next-line @typescript-eslint/require-await
    ): Promise<sourcegraph.ProgressReporter> => {
        // There's no guarantee that UI consumers have subscribed to the progress observable
        // by the time that an extension reports progress, so replay the latest report on subscription.
        const progressSubject = new ReplaySubject<sourcegraph.Progress>(1)

        // progress notifications have to be proxied since the observable
        // `progress` property cannot be cloned
        state.progressNotifications.next(
            proxy({
                baseNotification: {
                    message: options.title,
                    type: NotificationType.Log,
                },
                progress: proxySubscribable(progressSubject.asObservable()),
            })
        )

        // return ProgressReporter, which exposes a subset of Subject methods to extensions
        return {
            next: (progress: sourcegraph.Progress) => {
                progressSubject.next(progress)
            },
            error: (value: any) => {
                const error = asError(value)
                progressSubject.error({
                    message: error.message,
                    name: error.name,
                    stack: error.stack,
                })
            },
            complete: () => {
                progressSubject.complete()
            },
        }
    }

    // App
    const window: sourcegraph.Window = {
        get visibleViewComponents(): sourcegraph.ViewComponent[] {
            const entries = [...state.viewComponents.entries()]
            return sortBy(entries, 0).map(([, viewComponent]) => viewComponent)
        },
        get activeViewComponent(): sourcegraph.ViewComponent | undefined {
            return state.activeViewComponentChanges.value
        },
        activeViewComponentChanges: state.activeViewComponentChanges.asObservable(),
        showNotification: (message, type) => {
            state.plainNotifications.next({ message, type })
        },
        withProgress: async (options, task) => {
            const reporter = await createProgressReporter(options)
            try {
                const result = task(reporter)
                reporter.complete()
                return await result
            } catch (error) {
                reporter.error(error)
                throw error
            }
        },

        showProgress: options => createProgressReporter(options),
        showMessage: message => mainAPI.showMessage(message),
        showInputBox: options => mainAPI.showInputBox(options),
    }

    const app: Pick<
        typeof sourcegraph['app'],
        'activeWindow' | 'activeWindowChanges' | 'windows' | 'registerFileDecorationProvider'
    > = {
        // deprecated, add simple window getter
        get activeWindow() {
            return window
        },
        activeWindowChanges: new BehaviorSubject(window).asObservable(),
        get windows() {
            return [window]
        },
        registerFileDecorationProvider: (provider: sourcegraph.FileDecorationProvider): sourcegraph.Unsubscribable =>
            addWithRollback(state.fileDecorationProviders, provider),
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
    const registerDocumentHighlightProvider = (
        selector: sourcegraph.DocumentSelector,
        provider: sourcegraph.DocumentHighlightProvider
    ): sourcegraph.Unsubscribable => addWithRollback(state.documentHighlightProviders, { selector, provider })
    // definition
    const registerDefinitionProvider = (
        selector: sourcegraph.DocumentSelector,
        provider: sourcegraph.DefinitionProvider
    ): sourcegraph.Unsubscribable => addWithRollback(state.definitionProviders, { selector, provider })

    // GraphQL
    const graphQL: typeof sourcegraph['graphQL'] = {
        execute: (query, variables) => mainAPI.requestGraphQL(query, variables),
    }

    // Context + Contributions
    // Same implementation is exposed to main and extensions
    function updateContext(update: { [k: string]: unknown }): void {
        if (isMatch(state.context.value, update)) {
            return
        }
        const result: any = {}
        for (const [key, oldValue] of Object.entries(state.context.value)) {
            if (update[key] !== null) {
                result[key] = oldValue
            }
        }
        for (const [key, value] of Object.entries(update)) {
            if (value !== null) {
                result[key] = value
            }
        }
        state.context.next(result)
    }

    state.contributions.subscribe(entries => console.log('exthostcontributions', entries))

    wrapRemoteObservable(mainAPI.getEnabledExtensions()).subscribe(enabledExtensions =>
        console.log('enabled extensions exthost', { ms: Date.now() })
    )

    // TODO(tj): this is at the bottom since the `extensionAPI` object to be given
    // to extensions needs to be constructed first. Will need to migrate ALL services before actualizing this.
    ;(function handleExtensionActivation() {
        //     from(mainAPI.getScriptURLForExtension()).subscribe(getScriptURL => {
        //         const memoizedGetScriptURLForExtension = memoizeObservable<string, string | null>(
        //             url =>
        //                 // TODO: make getScriptURL batched to avoid N roundtrips on initial extension activation
        //                 (getScriptURL ? from(getScriptURL(url)) : of(url)).pipe(
        //                     catchError(error => {
        //                         console.error(`Error fetching extension script URL ${url}`, error)
        //                         return [null]
        //                     })
        //                 ),
        //             url => url
        //         )

        //         const enabledExtensions: Subscribable<ConfiguredExtension[]> = combineLatest([
        //             viewerConfiguredExtensions(),
        //             getSideloadedExtension(),
        //         ]).pipe(
        //             withLatestFrom(state.settings),
        //             map(([[configuredExtensions, sideloadedExtension], settings]) => {
        //                 let enabled = configuredExtensions.filter(extension =>
        //                     isExtensionEnabled(settings.final, extension.id)
        //                 )

        //                 if (sideloadedExtension) {
        //                     if (!isErrorLike(sideloadedExtension.manifest) && sideloadedExtension.manifest?.publisher) {
        //                         // Disable extension with the same ID while this extension is sideloaded
        //                         const constructedID = `${sideloadedExtension.manifest.publisher}/${sideloadedExtension.id}`
        //                         enabled = enabled.filter(extension => extension.id !== constructedID)
        //                     }

        //                     enabled.push(sideloadedExtension)
        //                 }
        //                 return enabled
        //             })
        //         )

        //         // TODO(tj): move to state
        //         const activatedExtensionIDs = new Set<string>()
        //         // An extension is activated when one or more of its activationEvents is true. After an extension has been
        //         // activated, it remains active for the rest of the session (i.e., for as long as the browser tab remains open)
        //         // as long as it remains enabled. If it is disabled, it is deactivated. (I.e., "activationEvents" are
        //         // retrospective/sticky.
        //         const activeExtensions: Subscribable<ExecutableExtension[]> = combineLatest([
        //             state.activeLanguages,
        //             enabledExtensions,
        //         ]).pipe(
        //             tap(([activeLanguages, enabledExtensions]) => {
        //                 console.log('checking active in host')

        //                 const activeExtensions = extensionsWithMatchedActivationEvent(enabledExtensions, activeLanguages)
        //                 for (const extension of activeExtensions) {
        //                     if (!activatedExtensionIDs.has(extension.id)) {
        //                         activatedExtensionIDs.add(extension.id)
        //                     }
        //                 }
        //             }),
        //             map(([, extensions]) =>
        //                 extensions ? extensions.filter(extension => activatedExtensionIDs.has(extension.id)) : []
        //             ),
        //             distinctUntilChanged((a, b) =>
        //                 isEqual(new Set(a.map(extension => extension.id)), new Set(b.map(extension => extension.id)))
        //             ),
        //             switchMap(extensions => {
        //                 console.log('in switchMap', { extensions })
        //                 return combineLatestOrDefault(
        //                     extensions.map(extension =>
        //                         memoizedGetScriptURLForExtension(getScriptURLFromExtensionManifest(extension)).pipe(
        //                             map(scriptURL =>
        //                                 scriptURL === null
        //                                     ? null
        //                                     : {
        //                                           id: extension.id,
        //                                           manifest: extension.manifest,
        //                                           scriptURL,
        //                                       }
        //                             )
        //                         )
        //                     )
        //                 )
        //             }),
        //             map(extensions => extensions.filter(isDefined)),
        //             distinctUntilChanged((a, b) =>
        //                 isEqual(new Set(a.map(extension => extension.id)), new Set(b.map(extension => extension.id)))
        //             ),
        //             tap(() => console.log('here in host'))
        //         )
        //         // TODO(tj): expose `activeExtensions` to main thread for global debug

        //         activeExtensions.subscribe(extensions => console.log({ extensions }))
        //     })

        const enabledExtensions: Subscribable<ConfiguredExtension[]> = combineLatest([
            viewerConfiguredExtensions(),
            getSideloadedExtension(),
        ]).pipe(
            withLatestFrom(state.settings),
            map(([[configuredExtensions, sideloadedExtension], settings]) => {
                let enabled = configuredExtensions.filter(extension => isExtensionEnabled(settings.final, extension.id))

                if (sideloadedExtension) {
                    if (!isErrorLike(sideloadedExtension.manifest) && sideloadedExtension.manifest?.publisher) {
                        // Disable extension with the same ID while this extension is sideloaded
                        const constructedID = `${sideloadedExtension.manifest.publisher}/${sideloadedExtension.id}`
                        enabled = enabled.filter(extension => extension.id !== constructedID)
                    }

                    enabled.push(sideloadedExtension)
                }
                return enabled
            })
        )
        // TODO: move activeExtensions to mainthread!
        // TODO(tj): move to state
        const activatedExtensionIDs = new Set<string>()
        const cachedScriptURLs = new Map<string, string>()
        // An extension is activated when one or more of its activationEvents is true. After an extension has been
        // activated, it remains active for the rest of the session (i.e., for as long as the browser tab remains open)
        // as long as it remains enabled. If it is disabled, it is deactivated. (I.e., "activationEvents" are
        // retrospective/sticky.
        const activeExtensions: Subscribable<ExecutableExtension[]> = combineLatest([
            state.activeLanguages,
            enabledExtensions,
        ]).pipe(
            tap(([activeLanguages, enabledExtensions]) => {
                console.log('checking active in host')

                const activeExtensions = extensionsWithMatchedActivationEvent(enabledExtensions, activeLanguages)
                for (const extension of activeExtensions) {
                    if (!activatedExtensionIDs.has(extension.id)) {
                        activatedExtensionIDs.add(extension.id)
                    }
                }
            }),
            map(([, extensions]) =>
                extensions ? extensions.filter(extension => activatedExtensionIDs.has(extension.id)) : []
            ),
            distinctUntilChanged((a, b) =>
                isEqual(new Set(a.map(extension => extension.id)), new Set(b.map(extension => extension.id)))
            ),
            switchMap(extensions => {
                console.log('in switchMap', { extensions })
                return combineLatestOrDefault(
                    extensions.map(extension =>
                        // create a Map of scriptURL -> scriptURL|blobURL for caching purposes.
                        // iterate over extensions, filter out extensions that have already resolved scriptURL.
                        // call `getScriptURLForExtension` return fn with new extensions if defined, then add those
                        // resolved script URLs (get id from index) to cache. this way we have one roundtrip at most
                        memoizedGetScriptURLForExtension(getScriptURLFromExtensionManifest(extension)).pipe(
                            map(scriptURL =>
                                scriptURL === null
                                    ? null
                                    : {
                                          id: extension.id,
                                          manifest: extension.manifest,
                                          scriptURL,
                                      }
                            )
                        )
                    )
                )
            }),
            map(extensions => extensions.filter(isDefined)),
            distinctUntilChanged((a, b) =>
                isEqual(new Set(a.map(extension => extension.id)), new Set(b.map(extension => extension.id)))
            ),
            tap(() => console.log('here in host'))
        )
        // TODO(tj): expose `activeExtensions` to main thread for global debug

        // activeExtensions.subscribe(extensions => console.log({ extensions }))

        // activate extensions

        // register extension contributions. try to find a way to prevent the original
        // "unsub existing on registration" problem.
        // maybe just remove support for observable entries! that's major indirection
        // since all we really want is:
        // activate extension -> fetch manifest -> register contributions (raw, will be parsed)
    })()

    // TJ note: try to more aggresively optimize now that more stuff is done on the
    // extension host side; minimize transfer weight!

    function viewerConfiguredExtensions() {
        return from(state.settings).pipe(
            map(settings => (settings.final?.extensions ? Object.keys(settings.final.extensions) : [])),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            switchMap(extensionIDs => queryConfiguredRegistryExtensions(graphQL, extensionIDs)),
            catchError(error => throwError(asError(error))),
            shareReplay(1)
            // TODO(tj); return to publish + refCount after fixing contrib problem
        )
    }

    function getSideloadedExtension(): Observable<ConfiguredExtension | null> {
        return wrapRemoteObservable(mainAPI.getSideloadedExtensionURL()).pipe(
            switchMap(url => (url ? getConfiguredSideloadedExtension(url) : of(null))),
            catchError(error => {
                console.error('Error sideloading extension', error)
                return of(null)
            })
        )
    }

    // consider how to avoid postMessage roundtrips for platforms that dont need this.
    // so I've reduced latency for platforms that don't need this, but how about for platforms that do?
    // batching?
    // NOW I've implemented batching, refactor
    const memoizedGetScriptURLForExtension = memoizeObservable<string, string | null>(
        url =>
            getScriptURL(null).pipe(
                switchMap(getScriptURL => (getScriptURL ? from(getScriptURL(url)) : of(url))),
                catchError(error => {
                    console.error(`Error fetching extension script URL ${url}`, error)
                    return [null]
                })
            ),
        url => url
    )
    // TODO(tj): explain why
    const getScriptURL = memoizeObservable(
        () => from(mainAPI.getScriptURLForExtension()),
        () => 'getScriptURL'
    )

    return {
        configuration: Object.assign(state.settings.pipe(mapTo(undefined)), {
            get: getConfiguration,
        }),
        exposedToMain,
        workspace,
        state,
        commands,
        search,
        languages: {
            registerHoverProvider,
            registerDocumentHighlightProvider,
            registerDefinitionProvider,
        },
        app,
        graphQL,
        internal: {
            updateContext,
        },
    }
}

// TODO (loic, felix) it might make sense to port tests with the rest of provider registries.
/**
 * Filt ers a list of Providers (P type) based on their selectors and a document
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
 * calls next() on behaviorSubject with a immutably added element ([...old, value])
 *
 * @param behaviorSubject subject that holds a collection
 * @param value to add to a collection
 * @returns Unsubscribable that will remove that element from the behaviorSubject.value and call next() again
 */
function addWithRollback<T>(behaviorSubject: BehaviorSubject<readonly T[]>, value: T): sourcegraph.Unsubscribable {
    behaviorSubject.next([...behaviorSubject.value, value])
    return {
        unsubscribe: () => behaviorSubject.next(behaviorSubject.value.filter(item => item !== value)),
    }
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
    return providersObservable
        .pipe(
            map(providers => filterProviders(providers)),

            switchMap(providers =>
                combineLatestOrDefault(
                    providers.map(provider =>
                        concat(
                            [LOADING],
                            providerResultToObservable(invokeProvider(provider)).pipe(
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

/**
 * Query the GraphQL API for registry metadata about the extensions given in {@link extensionIDs}.
 *
 * @returns An observable that emits once with the results.
 */
export function queryConfiguredRegistryExtensions(
    // TODO(tj): can copy this over to extension host, just replace platformContext.requestGraphQL
    // with mainThreadAPI.requestGraphQL
    graphQL: typeof sourcegraph['graphQL'],
    extensionIDs: string[]
): Observable<ConfiguredRegistryExtension[]> {
    if (extensionIDs.length === 0) {
        return of([])
    }
    const variables: GQL.IExtensionsOnExtensionRegistryArguments = {
        first: extensionIDs.length,
        prioritizeExtensionIDs: extensionIDs,
    }
    return from(
        graphQL.execute<GQL.IQuery, GQL.IExtensionsOnExtensionRegistryArguments>(
            gql`
                query Extensions($first: Int!, $prioritizeExtensionIDs: [String!]!) {
                    extensionRegistry {
                        extensions(first: $first, prioritizeExtensionIDs: $prioritizeExtensionIDs) {
                            nodes {
                                id
                                extensionID
                                url
                                manifest {
                                    raw
                                }
                                viewerCanAdminister
                            }
                        }
                    }
                }
            `,
            variables
        )
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.extensionRegistry?.extensions?.nodes) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.extensions.nodes.map(
                ({ id, extensionID, url, manifest, viewerCanAdminister }) => ({
                    id,
                    extensionID,
                    url,
                    manifest: manifest ? { raw: manifest.raw } : null,
                    viewerCanAdminister,
                })
            )
        }),
        map(registryExtensions => {
            const configuredExtensions: ConfiguredRegistryExtension[] = []
            for (const extensionID of extensionIDs) {
                const registryExtension = registryExtensions.find(extension => extension.extensionID === extensionID)
                configuredExtensions.push(
                    registryExtension
                        ? toConfiguredRegistryExtension(registryExtension)
                        : { id: extensionID, manifest: null, rawManifest: null, registryExtension: undefined }
                )
            }
            return configuredExtensions
        })
    )
}

/**
 * The manifest of an extension sideloaded during local development.
 *
 * Doesn't include {@link ExtensionManifest#url}, as this is added when
 * publishing an extension to the registry.
 * Instead, the bundle URL is computed from the manifest's `main` field.
 */
interface SideloadedExtensionManifest extends Omit<ExtensionManifest, 'url'> {
    name: string
    main: string
}

const getConfiguredSideloadedExtension = (baseUrl: string): Observable<ConfiguredExtension> =>
    fromFetch(`${baseUrl}/package.json`, { selector: response => checkOk(response).json() }).pipe(
        map(
            (response: SideloadedExtensionManifest): ConfiguredExtension => ({
                id: response.name,
                manifest: {
                    ...response,
                    url: `${baseUrl}/${response.main.replace('dist/', '')}`,
                },
                rawManifest: null,
            })
        )
    )

function extensionsWithMatchedActivationEvent(
    enabledExtensions: ConfiguredExtension[],
    visibleTextDocumentLanguages: ReadonlySet<string>
): ConfiguredExtension[] {
    const languageActivationEvents = new Set(
        [...visibleTextDocumentLanguages].map(language => `onLanguage:${language}`)
    )
    return enabledExtensions.filter(extension => {
        try {
            if (!extension.manifest) {
                const match = /^sourcegraph\/lang-(.*)$/.exec(extension.id)
                if (match) {
                    console.warn(
                        `Extension ${extension.id} has been renamed to sourcegraph/${match[1]}. It's safe to remove ${extension.id} from your settings.`
                    )
                } else {
                    console.warn(
                        `Extension ${extension.id} was not found. Remove it from settings to suppress this warning.`
                    )
                }
                return false
            }
            if (isErrorLike(extension.manifest)) {
                console.warn(extension.manifest)
                return false
            }
            if (!extension.manifest.activationEvents) {
                console.warn(`Extension ${extension.id} has no activation events, so it will never be activated.`)
                return false
            }
            return extension.manifest.activationEvents.some(
                event => event === '*' || languageActivationEvents.has(event)
            )
        } catch (error) {
            console.error(error)
        }
        return false
    })
}
