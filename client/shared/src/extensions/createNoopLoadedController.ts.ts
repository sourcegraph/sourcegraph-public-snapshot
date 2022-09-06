import { NEVER, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'

import { FlatExtensionHostAPI } from '../api/contract'
import { proxySubscribable } from '../api/extension/api/common'
import { createExtensionHostAPI } from '../api/extension/extensionHostApi'
import { createExtensionHostState } from '../api/extension/extensionHostState'
import { pretendRemote } from '../api/util'
import { newCodeIntelAPI } from '../codeintel/api'
import { CodeIntelContext, newSettingsGetter } from '../codeintel/legacy-extensions/api'
import { PlatformContext } from '../platform/context'
import { isSettingsValid } from '../settings/settings'

import { Controller } from './controller'
import { languageSpecs } from '../codeintel/legacy-extensions/language-specs/languages'

export function createNoopController(platformContext: PlatformContext): Controller {
    return {
        executeCommand: () => Promise.resolve(),
        commandErrors: NEVER,
        registerCommand: () => ({
            unsubscribe: () => {},
        }),
        extHostAPI: new Promise((resolve, reject) => {
            platformContext.settings.subscribe(async settingsCascade => {
                if (!isSettingsValid(settingsCascade)) {
                    reject(new Error('Settings are not valid'))
                    return
                }

                const extensionHostState = createExtensionHostState(
                    {
                        clientApplication: 'sourcegraph',
                        initialSettings: settingsCascade,
                    },
                    null,
                    null
                )
                const extensionHostAPI = injectNewCodeintel(createExtensionHostAPI(extensionHostState), {
                    requestGraphQL: platformContext.requestGraphQL,
                    telemetryService: platformContext.telemetryService,
                    settings: newSettingsGetter(platformContext.settings),
                })

                const implementationsLanguages = languageSpecs.filter(spec => spec.textDocumentImplemenationSupport)
                await extensionHostAPI.registerContributions({
                    actions: [
                        {
                            actionItem: {
                                description:
                                    '${!!config.codeIntel.mixPreciseAndSearchBasedReferences && "Hide search-based results when precise results are available" || ""}',
                                label:
                                    '${!!config.codeIntel.mixPreciseAndSearchBasedReferences && "Hide search-based results" || "Mix precise and search-based results"}',
                            },
                            command: 'updateConfiguration',
                            commandArguments: [
                                ['codeIntel.mixPreciseAndSearchBasedReferences'],
                                '${!config.codeIntel.mixPreciseAndSearchBasedReferences}',
                                null,
                                'json',
                            ],
                            id: 'mixPreciseAndSearchBasedReferences.toggle',
                            title:
                                '${!!config.codeIntel.mixPreciseAndSearchBasedReferences && "Hide search-based results when precise results are available" || "Mix precise and search-based results"}',
                        },
                        ...languageSpecs.map(spec => ({
                            actionItem: { label: 'Find implementations' },
                            command: 'open',
                            commandArguments: [
                                "${get(context, 'implementations_" +
                                    spec.languageID +
                                    "') && get(context, 'panel.url') && sub(get(context, 'panel.url'), 'panelID', 'implementations_" +
                                    spec.languageID +
                                    "') || 'noop'}",
                            ],
                            id: 'findImplementations',
                            title: 'Find implementations',
                        })),
                    ],
                    // configuration: {
                    //     properties: {
                    //         'basicCodeIntel.globalSearchesEnabled': {
                    //             description:
                    //                 'Whether to run global searches over all repositories. On instances with many repositories, this can lead to issues such as: low quality results, slow response times, or significant load on the Sourcegraph instance. Defaults to true.',
                    //             type: 'boolean',
                    //         },
                    //         'basicCodeIntel.includeArchives': {
                    //             description: 'Whether to include archived repositories in search results.',
                    //             type: 'boolean',
                    //         },
                    //         'basicCodeIntel.includeForks': {
                    //             description: 'Whether to include forked repositories in search results.',
                    //             type: 'boolean',
                    //         },
                    //         'basicCodeIntel.indexOnly': {
                    //             description: 'Whether to use only indexed requests to the search API.',
                    //             type: 'boolean',
                    //         },
                    //         'basicCodeIntel.unindexedSearchTimeout': {
                    //             description: 'The timeout (in milliseconds) for un-indexed search requests.',
                    //             type: 'number',
                    //         },
                    //         'codeIntel.disableRangeQueries': {
                    //             description: 'Whether to fetch multiple precise definitions and references on hover.',
                    //             type: 'boolean',
                    //         },
                    //         'codeIntel.disableSearchBased': {
                    //             description: 'Never fall back to search-based code intelligence.',
                    //             type: 'boolean',
                    //         },
                    //         'codeIntel.mixPreciseAndSearchBasedReferences': {
                    //             description: 'Whether to supplement precise references with search-based results.',
                    //             type: 'boolean',
                    //         },
                    //         'codeIntel.traceExtension': {
                    //             description: 'Whether to enable trace logging on the extension.',
                    //             type: 'boolean',
                    //         },
                    //     },
                    //     title: 'Search-based code intelligence settings',
                    // },
                    menus: {
                        hover: languageSpecs.map(spec => ({
                            action: 'findImplementations',
                            when:
                                "resource.language == '" +
                                spec.languageID +
                                "' && get(context, `implementations_${resource.language}`) && (goToDefinition.showLoading || goToDefinition.url || goToDefinition.error)",
                        })),
                        'panel/toolbar': [
                            {
                                action: 'mixPreciseAndSearchBasedReferences.toggle',
                                when: "panel.activeView.id == 'references' && !config.codeIntel.disableSearchBased",
                            },
                        ],
                    },
                })

                for (const spec of languageSpecs) {
                    if (spec.textDocumentImplemenationSupport) {
                        extensionHostAPI.updateContext({
                            [`implementations_${spec.languageID}`]: true,
                        })
                    }
                }

                // We don't have to load any extensions so we are already done
                extensionHostState.haveInitialExtensionsLoaded.next(true)

                resolve(pretendRemote(extensionHostAPI))
            })
        }),

        unsubscribe: () => {},
    }
}

// Replaces codeintel functions from the "old" extension/webworker extension API
// with new implementations of code that lives in this repository. The old
// implementation invoked codeintel functions via webworkers, and the codeintel
// implementation lived in a separate repository
// https://github.com/sourcegraph/code-intel-extensions Ideally, we should
// update all the usages of `comlink.Remote<FlatExtensionHostAPI>` with the new
// `CodeIntelAPI` interfaces, but that would require refactoring a lot of files.
// To minimize the risk of breaking changes caused by the deprecation of
// extensions, we monkey patch the old implementation with new implementations.
// The benefit of monkey patching is that we can optionally disable if for
// customers that choose to enable the legacy extensions.
export function injectNewCodeintel(
    old: FlatExtensionHostAPI,
    codeintelContext: CodeIntelContext
): FlatExtensionHostAPI {
    const codeintel = newCodeIntelAPI(codeintelContext)
    function thenMaybeLoadingResult<T>(promise: Observable<T>): Observable<MaybeLoadingResult<T>> {
        return promise.pipe(
            map(result => {
                const maybeLoadingResult: MaybeLoadingResult<T> = { isLoading: false, result }
                return maybeLoadingResult
            })
        )
    }

    const codeintelOverrides: Pick<
        FlatExtensionHostAPI,
        | 'providersForDocument'
        | 'getHover'
        | 'getDocumentHighlights'
        | 'getReferences'
        | 'getDefinition'
        | 'getLocations'
        | 'hasReferenceProvidersForDocument'
    > = {
        providersForDocument(textParameters) {
            return proxySubscribable(codeintel.providersForDocument(textParameters))
        },
        hasReferenceProvidersForDocument(textParameters) {
            return proxySubscribable(codeintel.hasReferenceProvidersForDocument(textParameters))
        },
        getLocations(id, parameters) {
            console.log({ id })
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getImplementations(parameters)))
        },
        getDefinition(parameters) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getDefinition(parameters)))
        },
        getReferences(parameters, context) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getReferences(parameters, context)))
        },
        getDocumentHighlights: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(codeintel.getDocumentHighlights(textParameters)),
        getHover: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(thenMaybeLoadingResult(codeintel.getHover(textParameters))),
    }

    return new Proxy(old, {
        get(target, prop) {
            const codeintelFunction = (codeintelOverrides as any)[prop]
            if (codeintelFunction) {
                return codeintelFunction
            }
            return Reflect.get(target, prop, ...arguments)
        },
    })
}
