import { ApolloClient, gql, throwServerError } from '@apollo/client'
import { isEqual } from 'lodash'
import { Observable, of, Subscription, from, ReplaySubject, Subscriber } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/graphql'

import { GetTemporarySettingsResult } from '../../graphql-operations'

import { TemporarySettings, TemporarySettingsSchema } from './TemporarySettings'

interface SettingInFlightResponse {
    loading: true
}

interface SettingFoundResponse<K extends keyof TemporarySettings, D extends TemporarySettings[K]> {
    loading: false
    value: TemporarySettingsSchema[K] | D
}

/**
 * Returned response when fetching an individual setting
 */
export type SettingResponse<K extends keyof TemporarySettings, D extends TemporarySettings[K]> =
    | SettingInFlightResponse
    | SettingFoundResponse<K, D>

/**
 * Returned response when fetching all settings
 */
interface TemporarySettingsResponse {
    loading: boolean
    settings: TemporarySettings
}

export class TemporarySettingsStorage {
    private settingsBackend: SettingsBackend = new LocalStorageSettingsBackend()
    // private settings: TemporarySettings = {}

    private onChange = new ReplaySubject<TemporarySettingsResponse>(1)

    private loadSubscription: Subscription | null = null
    private saveSubscription: Subscription | null = null

    public dispose(): void {
        this.loadSubscription?.unsubscribe()
        this.saveSubscription?.unsubscribe()
    }

    constructor(private apolloClient: ApolloClient<object> | null, isAuthenticatedUser: boolean) {
        if (isAuthenticatedUser) {
            if (!this.apolloClient) {
                throw new Error('Apollo-Client should be initialized for authenticated user')
            }

            this.setSettingsBackend(new ServersideSettingsBackend(this.apolloClient))
        } else {
            this.setSettingsBackend(new LocalStorageSettingsBackend())
        }
    }

    // This is public for testing purposes only so mocks can be provided.
    public setSettingsBackend(backend: SettingsBackend): void {
        this.loadSubscription?.unsubscribe()
        this.saveSubscription?.unsubscribe()

        this.settingsBackend = backend

        this.loadSubscription = this.settingsBackend.load().subscribe(({ loading, settings }) => {
            // TODO: Is it possible and safer to rely on Apollo cache update instead of manipulating settings?
            // Maybe we need to do this for localStorage
            // this.settings[key] = value
            this.onChange.next({ loading, settings })
        })
    }

    public set<K extends keyof TemporarySettings>(key: K, value: TemporarySettings[K]): void {
        // TODO: Is it possible and safer to rely on Apollo cache update instead of manipulating settings?
        // Maybe we need to do this for localStorage
        // this.settings[key] = value
        // this.onChange.next({ loading: false, value: this.settings })

        this.saveSubscription?.unsubscribe()
        this.saveSubscription = this.settingsBackend.edit({ [key]: value }).subscribe()
    }

    public get<K extends keyof TemporarySettings, D extends TemporarySettings[K]>(
        key: K,
        defaultValue: D
    ): Observable<SettingResponse<K, D>> {
        return this.onChange.pipe(
            map(({ loading, settings }) => {
                if (loading) {
                    return { loading: true } as const
                }

                return {
                    loading: false,
                    value: key in settings ? (settings[key] as TemporarySettingsSchema[K]) : defaultValue,
                } as const
            }),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )
    }
}

export interface SettingsBackend {
    load: () => Observable<TemporarySettingsResponse>
    edit: (settings: TemporarySettings) => Observable<void>
}

/**
 * Settings backend for unauthenticated users.
 * Settings are stored in `localStorage` and updated when
 * the `storage` event is fired on the window.
 */
class LocalStorageSettingsBackend implements SettingsBackend {
    private readonly TemporarySettingsKey = 'temporarySettings'

    public load(): Observable<TemporarySettingsResponse> {
        return new Observable<TemporarySettingsResponse>(observer => {
            let settingsLoaded = false

            const loadObserver = (observer: Subscriber<TemporarySettingsResponse>): void => {
                try {
                    const settings = localStorage.getItem(this.TemporarySettingsKey)
                    if (settings) {
                        const parsedSettings = JSON.parse(settings) as TemporarySettings
                        observer.next({ loading: false, settings: parsedSettings })
                        settingsLoaded = true
                    }
                } catch (error: unknown) {
                    console.error(error)
                }

                if (!settingsLoaded) {
                    observer.next({ loading: true, settings: {} })
                }
            }

            loadObserver(observer)

            const loadCallback = (): void => {
                loadObserver(observer)
            }

            window.addEventListener('storage', loadCallback)

            return () => {
                window.removeEventListener('storage', loadCallback)
            }
        })
    }

    public edit(newSettings: TemporarySettings): Observable<void> {
        try {
            const encodedCurrentSettings = localStorage.getItem(this.TemporarySettingsKey) || '{}'
            const currentSettings = JSON.parse(encodedCurrentSettings) as TemporarySettings
            localStorage.setItem(this.TemporarySettingsKey, JSON.stringify({ ...currentSettings, ...newSettings }))
        } catch (error: unknown) {
            console.error(error)
        }

        return of()
    }
}

/**
 * Settings backend for authenticated users that saves settings to the server.
 * Changes to settings are polled every 5 minutes.
 */
class ServersideSettingsBackend implements SettingsBackend {
    private readonly PollInterval = 1000 * 60 * 5 // 5 minutes

    private readonly GetTemporarySettingsQuery = gql`
        query GetTemporarySettings {
            temporarySettings {
                contents
            }
        }
    `

    private readonly EditTemporarySettingsMutation = gql`
        mutation EditTemporarySettings($contents: String!) {
            editTemporarySettings(settingsToEdit: $contents) {
                alwaysNil
            }
        }
    `

    constructor(private apolloClient: ApolloClient<object>) {}

    public load(): Observable<TemporarySettingsResponse> {
        const temporarySettingsQuery = this.apolloClient.watchQuery<GetTemporarySettingsResult>({
            query: this.GetTemporarySettingsQuery,
            pollInterval: this.PollInterval,
        })

        return fromObservableQuery(temporarySettingsQuery).pipe(
            map(({ data, loading }) => {
                let parsedSettings: TemporarySettings = {}

                try {
                    const settings = data.temporarySettings.contents
                    parsedSettings = JSON.parse(settings) as TemporarySettings
                } catch (error: unknown) {
                    console.error(error)
                }

                return { loading, settings: parsedSettings || {} }
            })
        )
    }

    public edit(newSettings: TemporarySettings): Observable<void> {
        try {
            const settingsString = JSON.stringify(newSettings)

            return from(
                this.apolloClient.mutate({
                    mutation: this.EditTemporarySettingsMutation,
                    variables: { contents: settingsString },
                    update: cache => {
                        const encodedCurrentSettings =
                            cache.readQuery<GetTemporarySettingsResult>({
                                query: this.GetTemporarySettingsQuery,
                            })?.temporarySettings.contents || '{}'
                        const currentSettings = JSON.parse(encodedCurrentSettings) as TemporarySettings

                        return cache.writeQuery<GetTemporarySettingsResult>({
                            query: this.GetTemporarySettingsQuery,
                            data: {
                                temporarySettings: {
                                    __typename: 'TemporarySettings',
                                    contents: JSON.stringify({ ...currentSettings, ...newSettings }),
                                },
                            },
                        })
                    },
                })
            ).pipe(
                map(() => {}) // Ignore return value, always empty
            )
        } catch (error: unknown) {
            console.error(error)
        }

        return of()
    }
}

/**
 * Static in memory setting backend for testing purposes
 */
export class InMemoryMockSettingsBackend implements SettingsBackend {
    constructor(private settings: TemporarySettings) {}
    public load(): Observable<TemporarySettingsResponse> {
        return of({ loading: false, settings: this.settings })
    }
    public edit(settings: TemporarySettings): Observable<void> {
        this.settings = { ...this.settings, ...settings }
        return of()
    }
}
