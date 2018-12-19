import { isEqual } from 'lodash'
import { BehaviorSubject, combineLatest, from, Observable, ObservableInput, of, Subscribable } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { catchError, distinctUntilChanged, flatMap, map, switchMap, tap } from 'rxjs/operators'
import {
    ConfiguredExtension,
    getScriptURLFromExtensionManifest,
    isExtensionEnabled,
} from '../../../extensions/extension'
import { viewerConfiguredExtensions } from '../../../extensions/helpers'
import { PlatformContext } from '../../../platform/context'
import { isErrorLike } from '../../../util/errors'
import { memoizeObservable } from '../../../util/memoizeObservable'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { Model } from '../model'
import { SettingsService } from './settings'

/**
 * The information about an extension necessary to execute and activate it.
 */
export interface ExecutableExtension extends Pick<ConfiguredExtension, 'id' | 'manifest'> {
    /** The URL to the JavaScript bundle of the extension. */
    scriptURL: string
}

const getConfiguredUnpackedExtension = async (baseUrl: string) => {
    const { response } = await ajax({
        url: `${baseUrl}/package.json`,
        responseType: 'json',
        crossDomain: true,
    }).toPromise()
    const ext: ConfiguredExtension = {
        id: response.name,
        manifest: {
            url: `${baseUrl}/${response.main.replace('dist/', '')}`,
            ...response,
        },
        rawManifest: null,
    }
    return ext
}

/**
 * Manages the set of extensions that are available and activated.
 *
 * @internal This is an internal implementation detail and is different from the product feature called the
 * "extension registry" (where users can search for and enable extensions).
 */
export class ExtensionsService {
    public constructor(
        private platformContext: Pick<PlatformContext, 'queryGraphQL' | 'getScriptURLForExtension'>,
        private model: Subscribable<Pick<Model, 'visibleViewComponents'>>,
        private settingsService: Pick<SettingsService, 'data'>,
        private extensionActivationFilter = extensionsWithMatchedActivationEvent
    ) {}

    protected configuredExtensions: Subscribable<ConfiguredExtension[]> = viewerConfiguredExtensions({
        settings: this.settingsService.data,
        queryGraphQL: this.platformContext.queryGraphQL,
    })

    /**
     * Returns an observable that emits the set of enabled extensions upon subscription and whenever it changes.
     *
     * Most callers should use {@link ExtensionsService#activeExtensions}.
     */
    private get enabledExtensions(): Subscribable<ConfiguredExtension[]> {
        return combineLatest(
            from(this.settingsService.data),
            from(this.configuredExtensions),
            from(this.unpackedExtension)
        ).pipe(
            flatMap(async ([settings, configuredExtensions, unpackedExtensionURL]) => {
                const enabled = [...configuredExtensions.filter(x => isExtensionEnabled(settings.final, x.id))]
                if (unpackedExtensionURL) {
                    try {
                        const unpackedExtension = await getConfiguredUnpackedExtension(unpackedExtensionURL)
                        enabled.push(unpackedExtension)
                    } catch (err) {
                        console.error(`Could not load unpacked extension from '${unpackedExtensionURL}': ${err}`)
                    }
                }
                return enabled
            })
        )
    }

    public unpackedExtension = new BehaviorSubject<string>('')

    /**
     * Returns an observable that emits the set of extensions that should be active, based on the previous and
     * current state and each available extension's activationEvents.
     *
     * An extension is activated when one or more of its activationEvents is true. After an extension has been
     * activated, it remains active for the rest of the session (i.e., for as long as the browser tab remains open)
     * as long as it remains enabled. If it is disabled, it is deactivated. (I.e., "activationEvents" are
     * retrospective/sticky.)
     *
     * @todo Consider whether extensions should be deactivated if none of their activationEvents are true (or that
     * plus a certain period of inactivity).
     *
     * @param extensionActivationFilter A function that returns the set of extensions that should be activated
     * based on the current model only. It does not need to account for remembering which extensions were
     * previously activated in prior states.
     */
    public get activeExtensions(): Subscribable<ExecutableExtension[]> {
        // Extensions that have been activated (including extensions with zero "activationEvents" that evaluate to
        // true currently).
        const activatedExtensionIDs: string[] = []
        return combineLatest(from(this.model), this.enabledExtensions).pipe(
            tap(([model, enabledExtensions]) => {
                const activeExtensions = this.extensionActivationFilter(enabledExtensions, model)
                for (const x of activeExtensions) {
                    if (!activatedExtensionIDs.includes(x.id)) {
                        activatedExtensionIDs.push(x.id)
                    }
                }
            }),
            map(([, extensions]) => (extensions ? extensions.filter(x => activatedExtensionIDs.includes(x.id)) : [])),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            switchMap(extensions =>
                combineLatestOrDefault(
                    extensions.map(x =>
                        this.memoizedGetScriptURLForExtension(getScriptURLFromExtensionManifest(x)).pipe(
                            map(scriptURL =>
                                scriptURL === null
                                    ? null
                                    : {
                                          id: x.id,
                                          manifest: x.manifest,
                                          scriptURL,
                                      }
                            )
                        )
                    )
                )
            ),
            map(extensions => extensions.filter((x): x is ExecutableExtension => x !== null))
        )
    }

    private memoizedGetScriptURLForExtension = memoizeObservable<string, string | null>(
        url =>
            asObservable(this.platformContext.getScriptURLForExtension(url)).pipe(
                catchError(err => {
                    console.error(`Error fetching extension script URL ${url}`, err)
                    return [null]
                })
            ),
        url => url
    )
}

function asObservable(input: string | ObservableInput<string>): Observable<string> {
    return typeof input === 'string' ? of(input) : from(input)
}

function extensionsWithMatchedActivationEvent(
    enabledExtensions: ConfiguredExtension[],
    model: Pick<Model, 'visibleViewComponents'>
): ConfiguredExtension[] {
    return enabledExtensions.filter(x => {
        try {
            if (!x.manifest) {
                console.warn(`Extension ${x.id} was not found. Remove it from settings to suppress this warning.`)
                return false
            } else if (isErrorLike(x.manifest)) {
                console.warn(x.manifest)
                return false
            } else if (!x.manifest.activationEvents) {
                console.warn(`Extension ${x.id} has no activation events, so it will never be activated.`)
                return false
            }
            const visibleTextDocumentLanguages = model.visibleViewComponents
                ? model.visibleViewComponents.map(({ item: { languageId } }) => languageId)
                : []
            return x.manifest.activationEvents.some(
                e => e === '*' || visibleTextDocumentLanguages.some(l => e === `onLanguage:${l}`)
            )
        } catch (err) {
            console.error(err)
        }
        return false
    })
}
