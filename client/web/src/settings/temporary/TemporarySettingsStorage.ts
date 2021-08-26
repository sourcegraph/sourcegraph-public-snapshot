import { ApolloClient, gql, NormalizedCacheObject } from '@apollo/client'
import { Observable, Subject, of, Subscription, from } from 'rxjs'
import { distinctUntilKeyChanged, map, startWith } from 'rxjs/operators'

import { AuthenticatedUser } from '../../auth'
import { GetTemporarySettingsResult } from '../../graphql-operations'

import { TemporarySettings } from './TemporarySettings'

export class TemporarySettingsStorage {
    private authenticatedUser: AuthenticatedUser | null = null
    private settingsBackend: SettingsBackend = new LocalStorageSettingsBackend()
    private settings: TemporarySettings = {}

    private onChange = new Subject<TemporarySettings>()

    private loadSubscription: Subscription | null = null
    private saveSubscription: Subscription | null = null

    public dispose(): void {
        this.loadSubscription?.unsubscribe()
        this.saveSubscription?.unsubscribe()
    }

    constructor(
        private apolloClient: ApolloClient<NormalizedCacheObject>,
        authenticatedUser: AuthenticatedUser | null
    ) {
        this.setAuthenticatedUser(authenticatedUser)
    }

    public setAuthenticatedUser(user: AuthenticatedUser | null): void {
        if (this.authenticatedUser !== user) {
            this.authenticatedUser = user

            if (this.authenticatedUser) {
                this.setSettingsBackend(new ServersideSettingsBackend(this.apolloClient))
            } else {
                this.setSettingsBackend(new LocalStorageSettingsBackend())
            }
        }
    }

    // This is public for testing purposes only so mocks can be provided.
    public setSettingsBackend(backend: SettingsBackend): void {
        this.loadSubscription?.unsubscribe()
        this.saveSubscription?.unsubscribe()

        this.settingsBackend = backend

        this.loadSubscription = this.settingsBackend.load().subscribe(settings => {
            this.settings = settings
            this.onChange.next(settings)
        })
    }

    public set<K extends keyof TemporarySettings>(key: K, value: TemporarySettings[K]): void {
        this.settings[key] = value
        this.onChange.next(this.settings)
        this.saveSubscription = this.settingsBackend.save(this.settings).subscribe()
    }

    public get<K extends keyof TemporarySettings>(key: K): Observable<TemporarySettings[K]> {
        return this.onChange.pipe(
            distinctUntilKeyChanged(key),
            map(settings => settings[key]),
            startWith(this.settings[key])
        )
    }
}

export interface SettingsBackend {
    load: () => Observable<TemporarySettings>
    save: (settings: TemporarySettings) => Observable<void>
}

class LocalStorageSettingsBackend implements SettingsBackend {
    private readonly TemporarySettingsKey = 'temporarySettings'

    public load(): Observable<TemporarySettings> {
        try {
            const settings = localStorage.getItem(this.TemporarySettingsKey)
            if (settings) {
                const parsedSettings = JSON.parse(settings) as TemporarySettings
                return of(parsedSettings)
            }
        } catch {
            // Ignore error
        }

        return of({})
    }

    public save(settings: TemporarySettings): Observable<void> {
        try {
            const settingsString = JSON.stringify(settings)
            localStorage.setItem(this.TemporarySettingsKey, settingsString)
        } catch {
            // Ignore error
        }

        return of()
    }
}

class ServersideSettingsBackend implements SettingsBackend {
    constructor(private apolloClient: ApolloClient<NormalizedCacheObject>) {}

    public load(): Observable<TemporarySettings> {
        const query = gql`
            query GetTemporarySettings {
                temporarySettings {
                    contents
                }
            }
        `

        return new Observable<TemporarySettings>(observer => {
            const subscription = this.apolloClient
                .watchQuery<GetTemporarySettingsResult>({ query })
                .subscribe({
                    next: result => {
                        let parsedSettings: TemporarySettings = {}
                        try {
                            const settings = result.data.temporarySettings.contents
                            parsedSettings = JSON.parse(settings) as TemporarySettings
                        } catch {
                            // Ignore error
                        }
                        if (parsedSettings) {
                            observer.next(parsedSettings)
                        } else {
                            observer.next({})
                        }
                    },
                    error: error => {
                        observer.error(error)
                    },
                    complete: () => {
                        observer.complete()
                    },
                })

            return () => subscription.unsubscribe()
        })
    }

    public save(settings: TemporarySettings): Observable<void> {
        const mutation = gql`
            mutation SetTemporarySettings($contents: String!) {
                overwriteTemporarySettings(contents: $contents) {
                    alwaysNil
                }
            }
        `

        try {
            const settingsString = JSON.stringify(settings)
            return from(this.apolloClient.mutate({ mutation, variables: { contents: settingsString } })).pipe(
                map(() => {}) // Ignore return value, always empty
            )
        } catch {
            // Ignore error
            return of()
        }
    }
}
