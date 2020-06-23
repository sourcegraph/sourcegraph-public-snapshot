import { SettingsCascade } from '../../settings/settings'
import { Remote, proxy } from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { BehaviorSubject, Subject, ReplaySubject, of, Observable, from, concat, isObservable } from 'rxjs'

import { FlatExtHostAPI, MainThreadAPI } from '../contract'
import { syncSubscription, isPromiseLike } from '../util'
import { switchMap, mergeMap, defaultIfEmpty, map, distinctUntilChanged, catchError } from 'rxjs/operators'
import { proxySubscribable } from './api/common'
import { combineLatestOrDefault } from '../../util/rxjs/combineLatestOrDefault'
import { LOADING } from '@sourcegraph/codeintellify'
import { fromHoverMerged } from '../client/types/hover'
import { isNot, isExactly } from '../../util/types'
import { isEqual } from 'lodash'

/**
 * Holds the entire state exposed to the extension host
 * as a single plain object
 */
export interface ExtState {
    settings: Readonly<SettingsCascade<object>>

    // Workspace
    roots: readonly sourcegraph.WorkspaceRoot[]
    versionContext: string | undefined

    // Search
    queryTransformers: sourcegraph.QueryTransformer[]

    // languages
    hoverProviders: sourcegraph.HoverProvider[]
    definitionProviders: sourcegraph.DefinitionProvider[]
    referenceProviders: sourcegraph.ReferenceProvider[]
}

export interface InitResult {
    configuration: sourcegraph.ConfigurationService
    workspace: PartialWorkspaceNamespace
    exposedToMain: FlatExtHostAPI
    // todo this is needed as a temp solution for getter problem
    state: Readonly<ExtState>
    commands: typeof sourcegraph['commands']
    search: typeof sourcegraph['search']
    languages: typeof sourcegraph['languages']
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
    initialSettings: Readonly<SettingsCascade<object>>
): InitResult => {
    const state: ExtState = {
        roots: [],
        versionContext: undefined,
        settings: initialSettings,
        queryTransformers: [],
        hoverProviders: [],
        definitionProviders: [],
        referenceProviders: [],
    }

    const configChanges = new BehaviorSubject<void>(undefined)
    // Most extensions never call `configuration.get()` synchronously in `activate()` to get
    // the initial settings data, and instead only subscribe to configuration changes.
    // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.

    const rootChanges = new Subject<void>()

    // Search
    const queryTransformersChanges = new ReplaySubject<sourcegraph.QueryTransformer[]>(1)
    queryTransformersChanges.next([])

    // Languages
    const hoverProviderChanges = new ReplaySubject<sourcegraph.HoverProvider[]>(1)
    hoverProviderChanges.next([])
    const definitionProviderChanges = new ReplaySubject<sourcegraph.DefinitionProvider[]>(1)
    definitionProviderChanges.next([])
    const referenceProviderChanges = new ReplaySubject<sourcegraph.ReferenceProvider[]>(1)
    referenceProviderChanges.next([])

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
                queryTransformersChanges.pipe(
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

        // Languages
        getHover: ({ textDocument, position }) =>
            proxySubscribable(
                hoverProviderChanges.pipe(
                    switchMap(providers =>
                        combineLatestOrDefault(
                            providers.map(provider => {
                                // TODO textDocument needs to be a TextDocument instance
                                const providerResult = provider.provideHover(textDocument, position)
                                return isPromiseLike(providerResult) || isObservable(providerResult)
                                    ? concat([LOADING], from(providerResult)).pipe(
                                          defaultIfEmpty(null),
                                          catchError(error => {
                                              console.error('Hover provider errored:', error)
                                              return [null]
                                          })
                                      )
                                    : of(providerResult)
                            })
                        ).pipe(
                            defaultIfEmpty<(typeof LOADING | sourcegraph.Hover | null | undefined)[]>([]),
                            map(hoversFromProviders => ({
                                isLoading: hoversFromProviders.some(hover => hover === LOADING),
                                result: fromHoverMerged(hoversFromProviders.filter(isNot(isExactly(LOADING)))),
                            })),
                            distinctUntilChanged((a, b) => isEqual(a, b))
                        )
                    )
                )
            ),
        getDefinitions: ({ textDocument, position }) =>
            proxySubscribable(
                // Do some stuff
                definitionProviderChanges.pipe()
            ),
        getReferences: ({ textDocument, position }) =>
            proxySubscribable(
                // Do some stuff
                definitionProviderChanges.pipe()
            ),
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
        registerQueryTransformer: transformer => {
            state.queryTransformers = state.queryTransformers.concat(transformer)
            queryTransformersChanges.next(state.queryTransformers)
            return {
                unsubscribe: () => {
                    // eslint-disable-next-line id-length
                    state.queryTransformers = state.queryTransformers.filter(t => t !== transformer)
                    queryTransformersChanges.next(state.queryTransformers)
                },
            }
        },
    }

    const languages: typeof sourcegraph['languages'] = {
        registerHoverProvider: (selector, provider) => {
            state.hoverProviders.push(provider)
            hoverProviderChanges.next(state.hoverProviders)
            return {
                unsubscribe: () => {
                    state.hoverProviders = state.hoverProviders.filter(
                        registeredProvider => registeredProvider !== provider
                    )
                    hoverProviderChanges.next(state.hoverProviders)
                },
            }
        },
        registerDefinitionProvider: (selector, provider) => {
            state.definitionProviders.push(provider)
            definitionProviderChanges.next(state.definitionProviders)
            return {
                unsubscribe: () => {
                    state.definitionProviders = state.definitionProviders.filter(
                        registeredProvider => registeredProvider !== provider
                    )
                    definitionProviderChanges.next(state.definitionProviders)
                },
            }
        },
        registerReferenceProvider: (selector, provider) => {
            state.referenceProviders.push(provider)
            referenceProviderChanges.next(state.referenceProviders)
            return {
                unsubscribe: () => {
                    state.referenceProviders = state.referenceProviders.filter(
                        registeredProvider => registeredProvider !== provider
                    )
                    referenceProviderChanges.next(state.referenceProviders)
                },
            }
        },
    }

    return {
        exposedToMain,
        configuration: Object.assign(configChanges.asObservable(), {
            get: getConfiguration,
        }),
        languages,
        workspace,
        state,
        commands,
        search,
    }
}
