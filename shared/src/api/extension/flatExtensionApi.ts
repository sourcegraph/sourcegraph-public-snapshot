import { SettingsCascade } from '../../settings/settings'
import { Remote, proxy } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { BehaviorSubject, Subject, of, Observable, from } from 'rxjs'
import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { syncSubscription } from '../util'
import { switchMap, mergeMap, map } from 'rxjs/operators'
import { proxySubscribable, providerResultToObservable } from './api/common'
import { TextDocumentIdentifier, match } from '../client/types/textDocument'
import { getModeFromPath } from '../../languages'
import { parseRepoURI } from '../../util/url'
import { ExtensionDocuments } from './api/documents'
import { toPosition } from './api/types'
import { TextDocumentPositionParams } from '../protocol'
import { ProvideTextDocumentHoverSignature, getHover } from './hover'

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
    hoverProviders: BehaviorSubject<RegisteredHoverProvider[]>
}

interface RegisteredHoverProvider {
    selector: sourcegraph.DocumentSelector
    provider: sourcegraph.HoverProvider
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: PartialWorkspaceNamespace
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
export type PartialWorkspaceNamespace = Omit<
    typeof sourcegraph['workspace'],
    'textDocuments' | 'onDidOpenTextDocument' | 'openedTextDocuments' | 'roots' | 'versionContext'
>
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
    textDocuments: ExtensionDocuments
): InitResult => {
    const state: ExtState = {
        roots: [],
        versionContext: undefined,
        settings: initialSettings,
        queryTransformers: new BehaviorSubject<sourcegraph.QueryTransformer[]>([]),
        hoverProviders: new BehaviorSubject<RegisteredHoverProvider[]>([]),
    }

    const configChanges = new BehaviorSubject<void>(undefined)
    // Most extensions never call `configuration.get()` synchronously in `activate()` to get
    // the initial settings data, and instead only subscribe to configuration changes.
    // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.

    const rootChanges = new Subject<void>()

    const versionContextChanges = new Subject<string | undefined>()

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
            const document = textDocuments.get(textParameters.textDocument.uri)

            const matchedProviders = state.hoverProviders.pipe(
                map(providers =>
                    providersForDocument(textParameters.textDocument, providers, ({ selector }) => selector).map(
                        ({ provider }): ProvideTextDocumentHoverSignature => parameters =>
                            providerResultToObservable(provider.provideHover(document, toPosition(parameters.position)))
                    )
                )
            )
            return proxySubscribable(getHover(matchedProviders, textParameters))
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
    const workspace: PartialWorkspaceNamespace = {
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
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        exposedToMain,
        workspace,
        state,
        commands,
        search,
        languages: {
            registerHoverProvider,
        },
    }
}

// TODO probably worth separate test suit.home
// maybe copy from registry.ts?
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

function addWithRollback<T>(behaviorSubject: BehaviorSubject<T[]>, value: T): sourcegraph.Unsubscribable {
    behaviorSubject.next([...behaviorSubject.value, value])
    return {
        unsubscribe: () => behaviorSubject.next(behaviorSubject.value.filter(p => p != value)),
    }
}
