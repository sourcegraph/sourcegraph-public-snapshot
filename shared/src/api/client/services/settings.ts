import { from, Subscribable } from 'rxjs'
import { first } from 'rxjs/operators'
import { PlatformContext } from '../../../platform/context'
import { isSettingsValid, Settings, SettingsCascadeOrError } from '../../../settings/settings'

/**
 * A key path that refers to a location in a JSON document.
 *
 * Each successive array element specifies an index in an object or array to descend into. For example, in the
 * object `{"a": ["x", "y"]}`, the key path `["a", 1]` refers to the value `"y"`.
 */
export type KeyPath = (string | number)[]

/**
 * An edit to apply to settings.
 */
export interface SettingsEdit {
    /** The key path to the value. */
    path: KeyPath

    /** The new value to insert at the key path. */
    value: any
}

/**
 * The settings service manages the settings cascade for the viewer.
 *
 * @template S The settings type.
 */
export interface SettingsService<S extends Settings = Settings> {
    /**
     * The settings cascade.
     */
    data: Subscribable<SettingsCascadeOrError<S>>

    /**
     * Update the settings for the settings subject with the highest precedence.
     *
     * @todo Support specifying which settings subject whose settings to update.
     */
    update(edit: SettingsEdit): Promise<void>
}

/**
 * Create a {@link SettingsService} instance.
 *
 * @template S The settings type.
 */
export function createSettingsService<S extends Settings = Settings>({
    settings: data,
    updateSettings,
}: Pick<PlatformContext, 'settings' | 'updateSettings'>): SettingsService<S> {
    return {
        data: data as Subscribable<SettingsCascadeOrError<S>>, // cast to add type parameter S
        update: async edit => {
            const settings = await from(data)
                .pipe(first())
                .toPromise()
            if (!isSettingsValid(settings)) {
                throw new Error('invalid settings (internal error)')
            }
            const subject = settings.subjects[settings.subjects.length - 1]
            await updateSettings(subject.subject.id, edit)
        },
    }
}
