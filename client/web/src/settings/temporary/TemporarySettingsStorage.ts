import { Observable, Subject, of, Subscription } from 'rxjs'
import { distinctUntilKeyChanged, map, startWith } from 'rxjs/operators'

import { AuthenticatedUser } from '../../auth'

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

    public setAuthenticatedUser(user: AuthenticatedUser | null): void {
        if (this.authenticatedUser !== user) {
            this.authenticatedUser = user

            if (this.authenticatedUser) {
                // This will change to GraphQL backend in a future change
                this.setSettingsBackend(new LocalStorageSettingsBackend())
            } else {
                this.setSettingsBackend(new LocalStorageSettingsBackend())
            }

            this.loadSubscription = this.settingsBackend.load().subscribe(settings => {
                this.settings = settings
                this.onChange.next(settings)
            })
        }
    }

    // This is public for testing purposes only so mocks can be provided.
    public setSettingsBackend(backend: SettingsBackend): void {
        console.log(`TemporarySettingsStorage: Setting settings backend to ${backend.constructor.name}`)

        this.loadSubscription?.unsubscribe()
        this.saveSubscription?.unsubscribe()

        this.settingsBackend = backend

        this.loadSubscription = this.settingsBackend.load().subscribe(settings => {
            console.log(`TemporarySettingsStorage: Loaded settings: ${JSON.stringify(settings)}`)

            this.settings = settings
            this.onChange.next(settings)
        })
    }

    public set<K extends keyof TemporarySettings>(key: K, value: TemporarySettings[K]): void {
        console.log(`TemporarySettingsStorage: Setting ${key} to ${JSON.stringify(value)}`)

        this.settings[key] = value
        this.onChange.next(this.settings)
        this.saveSubscription = this.settingsBackend.save(this.settings).subscribe()
    }

    public get<K extends keyof TemporarySettings>(key: K): Observable<TemporarySettings[K]> {
        console.log(`TemporarySettingsStorage: Getting ${key} with initial value ${JSON.stringify(this.settings[key])}`)

        return this.onChange.pipe(
            distinctUntilKeyChanged(key),
            map(settings => {
                console.log(
                    `TemporarySettingsStorage: Getting ${key} from event with value ${JSON.stringify(settings[key])}`
                )
                return settings[key]
            }),
            startWith(this.settings[key])
        )
    }
}

interface SettingsBackend {
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
