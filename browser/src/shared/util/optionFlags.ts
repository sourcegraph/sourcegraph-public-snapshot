export type OptionFlagKey =
    | 'sendTelemetry'
    | 'allowErrorReporting'
    | 'experimentalLinkPreviews'
    | 'experimentalTextFieldCompletion'

export interface OptionFlagWithValue {
    label: string
    key: OptionFlagKey
    value: boolean
    hidden?: boolean
}
