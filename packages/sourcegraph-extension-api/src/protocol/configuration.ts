/**
 * A key path that refers to a location in a JSON document.
 *
 * Each successive array element specifies an index in an object or array to descend into. For example, in the
 * object `{"a": ["x", "y"]}`, the key path `["a", 1]` refers to the value `"y"`.
 */
export type KeyPath = (string | number)[]

/** The parameters for the configuration/update request. */
export interface ConfigurationUpdateParams {
    /** The key path to the value. */
    path: KeyPath

    /** The new value to insert at the key path. */
    value: any
}

/**
 * The configuration cascade, which describes the configuration at multiple levels (subjects), depending on the
 * client. The merged configuration is the result of shallow-merging configuration objects from all subjects, in
 * order from lower to higher precedence.
 *
 * For example, the client might support configuring settings globally and per-user, and it is designed so that
 * user settings override global settings. Then there would be two subjects, one for global settings and one for
 * the user.
 *
 * @template S the settings type
 */
export interface ConfigurationCascade<C = any> {
    /** The final settings, merged from all subjects in the cascade. */
    merged: C

    /**
     * The configuration subjects in the cascade, from lower to higher precedence.
     *
     * Extensions: The merged settings value usually suffices.
     */
    subjects?: { settings: C }[]
}
