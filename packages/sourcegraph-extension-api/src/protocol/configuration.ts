/**
 * A key path that refers to a location in a JSON document.
 *
 * Each successive array element specifies an index in an object or array to descend into. For example, in the
 * object `{"a": ["x", "y"]}`, the key path `["a", 1]` refers to the value `"y"`.
 */
export type KeyPath = (string | number)[]

export interface ConfigurationUpdateParams {
    /** The key path to the value. */
    path: KeyPath

    /** The new value to insert at the key path. */
    value: any
}

/**
 * The settings cascade, which describes the settings at multiple levels (subjects), depending on the client. The
 * merged settings is the result of shallow-merging settings objects from all subjects, in order from lower to
 * higher precedence.
 *
 * For example, the client might support settings globally and per-user, and it is designed so that
 * user settings override global settings. Then there would be two subjects, one for global settings and one for
 * the user.
 *
 * @template S the settings type
 */
export interface SettingsCascade<C = any> {
    /** The final settings, merged from all subjects in the cascade. */
    final: C

    /**
     * The settings subjects in the cascade, from lower to higher precedence.
     *
     * Extensions: The merged settings value usually suffices.
     */
    subjects?: { settings: C }[]
}
