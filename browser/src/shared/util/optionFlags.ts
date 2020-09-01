import { observeStorageKey } from '../../browser-extension/web-extension-api/storage'
import { map, distinctUntilChanged } from 'rxjs/operators'
import { Observable } from 'rxjs'

const OPTION_FLAGS_SYNC_STORAGE_KEY = 'featureFlags'

export type OptionFlagKey =
    | 'sendTelemetry'
    | 'allowErrorReporting'
    | 'experimentalLinkPreviews'
    | 'experimentalTextFieldCompletion'

export interface OptionFlagDefinition {
    label: string
    key: OptionFlagKey
    hidden?: boolean
}

export interface OptionFlagWithValue<T = boolean> extends OptionFlagDefinition {
    value: T
}

export type OptionFlagValues<T = boolean> = Record<OptionFlagKey, T>

export const optionFlagDefinitions: OptionFlagDefinition[] = [
    {
        key: 'sendTelemetry',
        label: 'Send telemetry',
    },
    {
        key: 'allowErrorReporting',
        label: 'Allow error reporting',
    },
    {
        key: 'experimentalLinkPreviews',
        label: 'Experimental link previews',
    },
    {
        key: 'experimentalTextFieldCompletion',
        label: 'Experimental text field completion',
    },
]

const optionFlagDefaults: OptionFlagValues = {
    sendTelemetry: false,
    allowErrorReporting: false,
    experimentalLinkPreviews: false,
    experimentalTextFieldCompletion: false,
}

export function assignOptionFlagValues(values: OptionFlagValues): OptionFlagWithValue[] {
    return optionFlagDefinitions.map(flag => ({ ...flag, value: values[flag.key] }))
}

export function applyOptionFlagDefaults(values: Partial<OptionFlagValues> | undefined): OptionFlagValues {
    return { ...optionFlagDefaults, ...values }
}
export const observeOptionFlags = (): Observable<OptionFlagValues> =>
    observeStorageKey('sync', OPTION_FLAGS_SYNC_STORAGE_KEY).pipe(map(applyOptionFlagDefaults))

export function observeOptionFlag(key: OptionFlagKey): Observable<boolean | undefined> {
    return observeOptionFlags().pipe(
        map(value => value?.[key]),
        distinctUntilChanged()
    )
}
