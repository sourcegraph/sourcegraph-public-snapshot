import { Remote } from 'comlink'
import { BehaviorSubject, combineLatest, from, Observable, Subscription } from 'rxjs'
import { catchError, concatMap, distinctUntilChanged, map, tap } from 'rxjs/operators'

import { ConfiguredExtension, getScriptURLFromExtensionManifest, splitExtensionID } from '../../extensions/extension'
import { areExtensionsSame, getEnabledExtensionsForSubject } from '../../extensions/extensions'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { hashCode } from '../../util/hashCode'
import { memoizeObservable } from '../../util/memoizeObservable'
import { wrapRemoteObservable } from '../client/api/common'
import { MainThreadAPI } from '../contract'
import { Contributions } from '../protocol'
import { tryCatchPromise } from '../util'

import { parseContributionExpressions } from './api/contribution'
import { ExtensionHostState } from './extensionHostState'

export function observeActiveExtensions(
    mainAPI: Remote<MainThreadAPI>
): {
    activeLanguages: ExtensionHostState['activeLanguages']
    activeExtensions: ExtensionHostState['activeExtensions']
} {
    const activeLanguages = new BehaviorSubject<ReadonlySet<string>>(new Set())
    const enabledExtensions = wrapRemoteObservable(mainAPI.getEnabledExtensions())
    const activatedExtensionIDs = new Set<string>()

    const activeExtensions: Observable<(ConfiguredExtension | ExecutableExtension)[]> = combineLatest([
        activeLanguages,
        enabledExtensions,
    ]).pipe(
        tap(([activeLanguages, enabledExtensions]) => {
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
        distinctUntilChanged((a, b) => areExtensionsSame(a, b))
    )

    return {
        activeLanguages,
        activeExtensions,
    }
}

export function activateExtensions(
    state: Pick<ExtensionHostState, 'activeExtensions' | 'contributions' | 'haveInitialExtensionsLoaded' | 'settings'>,
    mainAPI: Remote<Pick<MainThreadAPI, 'getScriptURLForExtension' | 'logEvent'>>,
    /**
     * Function that activates an extension.
     * Returns a promise that resolves once the extension is activated.
     * */
    activate = activateExtension,
    /**
     * Function that de-activates an extension.
     * Returns a promise that resolves once the extension is de-activated.
     * */
    deactivate = deactivateExtension
): Subscription {
    const getScriptURLs = memoizeObservable(
        () =>
            from(mainAPI.getScriptURLForExtension()).pipe(
                map(getScriptURL => {
                    function getBundleURLs(urls: string[]): Promise<(string | ErrorLike)[]> {
                        return getScriptURL ? getScriptURL(urls) : Promise.resolve(urls)
                    }

                    return getBundleURLs
                })
            ),
        () => 'getScriptURL'
    )

    const previouslyActivatedExtensions = new Set<string>()
    const extensionContributions = new Map<string, Contributions>()
    const contributionsToAdd = new Map<string, Contributions>()
    const extensionsSubscription = combineLatest([state.activeExtensions, getScriptURLs(null)])
        .pipe(
            concatMap(([activeExtensions, getScriptURLs]) => {
                const toDeactivate = new Set<string>()
                const toActivate = new Map<string, ConfiguredExtension | ExecutableExtension>()
                const activeExtensionIDs = new Set<string>()

                for (const extension of activeExtensions) {
                    // Populate set of currently active extension IDs
                    activeExtensionIDs.add(extension.id)
                    // Populate map of extensions to activate
                    if (!previouslyActivatedExtensions.has(extension.id)) {
                        toActivate.set(extension.id, extension)
                    }
                }

                for (const id of previouslyActivatedExtensions) {
                    // Populate map of extensions to deactivate
                    if (!activeExtensionIDs.has(id)) {
                        toDeactivate.add(id)
                    }
                }

                return from(
                    getScriptURLs(
                        [...toActivate.values()].map(extension => {
                            if ('scriptURL' in extension) {
                                // This is already an executable extension (inline extension)
                                return extension.scriptURL
                            }

                            return getScriptURLFromExtensionManifest(extension)
                        })
                    ).then(scriptURLs => {
                        // TODO: (not urgent) add scriptURL cache

                        const executableExtensionsToActivate: ExecutableExtension[] = [...toActivate.values()]
                            .map((extension, index) => ({
                                id: extension.id,
                                manifest: extension.manifest,
                                scriptURL: scriptURLs[index],
                            }))
                            .filter(
                                (extension): extension is ExecutableExtension => typeof extension.scriptURL === 'string'
                            )

                        return { toActivate: executableExtensionsToActivate, toDeactivate }
                    })
                ).pipe(
                    tap(({ toActivate }) => {
                        for (const extension of toActivate) {
                            if (
                                extension.manifest &&
                                !isErrorLike(extension.manifest) &&
                                extension.manifest.contributes
                            ) {
                                const parsedContributions = parseContributionExpressions(extension.manifest.contributes)
                                extensionContributions.set(extension.id, parsedContributions)
                                // Extension contributions additions and removals are batched
                                contributionsToAdd.set(extension.id, parsedContributions)
                            }
                        }
                    }),
                    map(({ toActivate, toDeactivate }) => {
                        // We could log the event after the activation promise resolves to ensure that there wasn't
                        // an error during activation, but we want to track the maximum number of times an extension could have been useful.
                        // Since extension activation is passive from the user's perspective, and we don't yet track extension usage events,
                        // there's no way that we could measure how often extensions are actually useful anyways.
                        const defaultExtensions =
                            getEnabledExtensionsForSubject(state.settings.value, 'DefaultSettings') || {}

                        return from(
                            Promise.all([
                                toActivate.map(({ id, scriptURL }) => {
                                    console.log(`Activating Sourcegraph extension: ${id}`)

                                    // We only want to log non-default extension events
                                    if (!defaultExtensions[id]) {
                                        // Hash extension IDs that specify host, since that means that it's a private registry extension.
                                        const telemetryExtensionID = splitExtensionID(id).host ? hashCode(id, 20) : id
                                        mainAPI
                                            .logEvent('ExtensionActivation', {
                                                extension_id: telemetryExtensionID,
                                            })
                                            .catch(() => {
                                                // noop
                                            })
                                    }

                                    return activate(id, scriptURL).catch(error =>
                                        console.error(`Error activating extension ${id}:`, asError(error))
                                    )
                                }),
                                [...toDeactivate].map(id =>
                                    deactivate(id).catch(error =>
                                        console.error(`Error deactivating extension ${id}:`, asError(error))
                                    )
                                ),
                            ])
                        )
                    }),
                    map(() => ({ activated: toActivate, deactivated: toDeactivate })),
                    catchError(error => {
                        console.error('Uncaught error during extension activation', error)
                        return []
                    })
                )
            })
        )
        .subscribe(({ activated, deactivated }) => {
            const contributionsToRemove = [...deactivated].map(id => extensionContributions.get(id)).filter(Boolean)

            for (const id of deactivated) {
                previouslyActivatedExtensions.delete(id)
                extensionContributions.delete(id)
            }

            for (const [id] of activated) {
                previouslyActivatedExtensions.add(id)
            }

            if (contributionsToAdd.size > 0) {
                state.contributions.next([...state.contributions.value, ...contributionsToAdd.values()])
                contributionsToAdd.clear()
            }

            if (contributionsToRemove.length > 0) {
                state.contributions.next(
                    state.contributions.value.filter(contributions => !contributionsToRemove.includes(contributions))
                )
            }

            if (state.haveInitialExtensionsLoaded.value === false) {
                state.haveInitialExtensionsLoaded.next(true)
            }
        })

    return extensionsSubscription
}

/** The WebWorker's global scope */
declare const self: any

/** Extensions' deactivate functions. */
const extensionDeactivates = new Map<string, () => void | Promise<void>>()

async function activateExtension(extensionID: string, bundleURL: string): Promise<void> {
    // Load the extension bundle and retrieve the extension entrypoint module's exports on
    // the global `module` property.
    try {
        const exports = {}
        self.exports = exports
        self.module = { exports }
        self.importScripts(bundleURL)
    } catch (error) {
        throw new Error(
            `error thrown while executing extension ${JSON.stringify(
                extensionID
            )} bundle (in importScripts of ${bundleURL}): ${String(error)}`
        )
    }
    const extensionExports = self.module.exports
    delete self.exports
    delete self.module

    if (!('activate' in extensionExports)) {
        throw new Error(
            `extension bundle for ${JSON.stringify(extensionID)} has no exported activate function (in ${bundleURL})`
        )
    }

    // During extension deactivation, we first call the extension's `deactivate` function and then unsubscribe
    // the Subscription passed to the `activate` function.
    const extensionSubscriptions = new Subscription()
    const extensionDeactivate = typeof extensionExports.deactivate === 'function' ? extensionExports.deactivate : null
    extensionDeactivates.set(extensionID, async () => {
        try {
            if (extensionDeactivate) {
                await Promise.resolve(extensionDeactivate())
            }
        } finally {
            extensionSubscriptions.unsubscribe()
        }
    })

    // The behavior should be consistent for both sync and async activate functions that throw
    // errors or reject. Both cases should yield a rejected promise.
    //
    // TODO(sqs): Add timeouts to prevent long-running activate or deactivate functions from
    // significantly delaying other extensions.
    await tryCatchPromise<void>(() => extensionExports.activate({ subscriptions: extensionSubscriptions })).catch(
        error => {
            error = asError(error)
            throw Object.assign(
                new Error(
                    `error during extension ${JSON.stringify(extensionID)} activate function: ${String(
                        error.stack || error
                    )}`
                ),
                { error }
            )
        }
    )
}

async function deactivateExtension(extensionID: string): Promise<void> {
    const deactivate = extensionDeactivates.get(extensionID)
    if (deactivate) {
        extensionDeactivates.delete(extensionID)
        await Promise.resolve(deactivate())
    }
}

export function extensionsWithMatchedActivationEvent(
    enabledExtensions: (ConfiguredExtension | ExecutableExtension)[],
    visibleTextDocumentLanguages: ReadonlySet<string>
): (ConfiguredExtension | ExecutableExtension)[] {
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

/**
 * The information about an extension necessary to execute and activate it.
 */
export interface ExecutableExtension extends Pick<ConfiguredExtension, 'id' | 'manifest'> {
    /** The URL to the JavaScript bundle of the extension. */
    scriptURL: string
}
