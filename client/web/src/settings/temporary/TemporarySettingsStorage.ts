import { Observable, Subject } from 'rxjs'
import { distinctUntilKeyChanged, map } from 'rxjs/operators'

import { AuthenticatedUser } from '../../auth'

import { TemporarySettings } from './TemporarySettings'

export class TemporarySettingsStorage {
    private authenticatedUser: AuthenticatedUser | null = null
    private settings: TemporarySettings = {}

    private onChange = new Subject<TemporarySettings>()

    public setAuthenticatedUser(user: AuthenticatedUser | null): void {
        this.authenticatedUser = user
    }

    public get<K extends keyof TemporarySettings>(key: K): TemporarySettings[K] {
        return this.settings[key]
    }

    public set<K extends keyof TemporarySettings>(key: K, value: TemporarySettings[K]): void {
        this.settings[key] = value
        this.onChange.next(this.settings)
    }

    public listen<K extends keyof TemporarySettings>(key: K): Observable<TemporarySettings[K]> {
        return this.onChange.pipe(
            distinctUntilKeyChanged(key),
            map(settings => settings[key])
        )
    }
}
