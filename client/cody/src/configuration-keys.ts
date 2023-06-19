import { camelCase } from 'lodash'

import packageJson from '../package.json'

const { properties } = packageJson.contributes.configuration

type ConfigurationKeysMap = {
    // Use key remapping to get a nice typescript interface with the correct keys.
    // https://www.typescriptlang.org/docs/handbook/2/mapped-types.html#key-remapping-via-as
    [key in keyof typeof properties as RemoveCodyPrefixAndCamelCase<key>]: key
}

function getConfigFromPackageJson(): ConfigurationKeysMap {
    return Object.keys(properties).reduce<ConfigurationKeysMap>((acc, key) => {
        // Remove the `cody.` prefix and camelCase the rest of the key.
        const keyProperty = camelCase(key.split('.').slice(1).join('.')) as keyof ConfigurationKeysMap

        // This is just to hard to type correctly 😜 and it's doesn't make any difference.
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        acc[keyProperty] = key as ConfigurationKeysMap[typeof keyProperty]

        return acc
    }, {} as ConfigurationKeysMap)
}

// Use template literal type for string manipulation. See examples here:
// https://www.typescriptlang.org/docs/handbook/2/template-literal-types.html
type RemoveCodyPrefixAndCamelCase<T extends string> = T extends `cody.${infer A}`
    ? A extends `${infer B}.${infer C}`
        ? `${B}${CamelCaseDotSeparatedFragments<C>}`
        : `${A}`
    : never

type CamelCaseDotSeparatedFragments<T extends string> = T extends `${infer A}.${infer B}`
    ? `${Capitalize<A>}${CamelCaseDotSeparatedFragments<B>}`
    : `${Capitalize<T>}`

/**
 * Automatically infer the configuration keys from the package.json in a type-safe way.
 * All the keys are mapped into the `CONFIG_KEY` object by removing the `cody.` prefix and
 * camelcasing the rest of the dot separated fragments.
 *
 * We should avoid specifiying config keys manually and instead rely on constant.
 * No manual changes will be required in this file when changing configuration keys in package.json.
 * Typescript will error for all outdated/missing keys.
 */
export const CONFIG_KEY = getConfigFromPackageJson()

export type ConfigKeys = keyof typeof CONFIG_KEY
