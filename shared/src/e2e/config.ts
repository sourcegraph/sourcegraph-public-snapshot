import { pick } from 'lodash'

/**
 * Defines configuration for e2e tests. This is as-yet incomplete as some config
 * depended on by other modules is not included here.
 */
export interface Config {
    sudoToken: string
    sudoUsername: string
    gitHubToken: string
    sourcegraphBaseUrl: string
    includeAdminOnboarding: boolean
    testUserPassword: string
}

interface Field<T = string> {
    envVar: string
    description?: string
    defaultValue?: T
}

interface FieldParser<T = string> {
    parser: (rawValue: string) => T
}

type ConfigFields = {
    [K in keyof Config]: Field<Config[K]> & (Config[K] extends string ? Partial<FieldParser> : FieldParser<Config[K]>)
}

const parseBool = (s: string): boolean => {
    if (['1', 't', 'true'].some(v => v.toLowerCase() === s)) {
        return true
    }
    if (['0', 'f', 'false'].some(v => v.toLowerCase() === s)) {
        return false
    }
    throw new Error(`could not parse string ${JSON.stringify(s)} to boolean`)
}

const parseString = (s: string): string => s

const configFields: ConfigFields = {
    sudoToken: {
        envVar: 'SOURCEGRAPH_SUDO_TOKEN',
        parser: parseString,
        description:
            'An access token with "site-admin:sudo" permissions. This will be used to impersonate users in requests.',
    },
    sudoUsername: {
        envVar: 'SOURCEGRAPH_SUDO_USER',
        parser: parseString,
        description: 'The site-admin-level username that will be impersonated with the sudo access token.',
    },
    gitHubToken: {
        envVar: 'GITHUB_TOKEN',
        parser: parseString,
        description: 'A GitHub token that will be used to authenticate a GitHub external service.',
    },
    sourcegraphBaseUrl: {
        envVar: 'SOURCEGRAPH_BASE_URL',
        parser: parseString,
        defaultValue: 'http://localhost:3080',
        description:
            'The base URL of the Sourcegraph instance, e.g., https://sourcegraph.sgdev.org or http://localhost:3080.',
    },
    includeAdminOnboarding: {
        envVar: 'INCLUDE_ADMIN_ONBOARDING',
        parser: parseBool,
        description:
            'If true, include admin onboarding tests, which assume none of the admin onboarding steps have yet completed on the instance. If those steps have already been completed, this test will fail.',
    },
    testUserPassword: {
        envVar: 'TEST_USER_PASSWORD',
        parser: parseString,
        description:
            'The password to use for any test users that are created. This password should be secure and unguessable when running against production Sourcegraph instances.',
    },
}

/**
 * Reads e2e config from environment variables. The caller should specify the config fields that it
 * depends on.
 */
export function getConfig<T extends keyof Config>(...required: T[]): Pick<Config, T> {
    // Read config
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const config: { [key: string]: any } = {}
    for (const fieldName of required) {
        const field = configFields[fieldName]
        if (field.defaultValue !== undefined) {
            config[fieldName] = field.defaultValue
        }
        const envValue = process.env[field.envVar]
        if (envValue) {
            config[fieldName] = field.parser ? field.parser(envValue) : envValue
        }
    }

    // Check required fields for type safety
    const missingKeys = required.filter(key => config[key] === undefined)
    if (missingKeys.length > 0) {
        const fieldInfo = (k: T): string => {
            const field = configFields[k]
            if (!field) {
                return ''
            }
            const info = [field.envVar]
            if (field.defaultValue) {
                info.push(`default value: ${field.defaultValue}`)
            }
            if (field.description) {
                info.push(`description: ${field.description}`)
            }
            return `${info.join(', ')}`
        }
        throw new Error(`FAIL: Required config was not provided. These environment variables were missing:

${missingKeys.map(k => `- ${fieldInfo(k)}`).join('\n')}

The recommended way to set them is to install direnv (https://direnv.net) and
create a .envrc file at the root of this repository.
        `)
    }

    return pick(config, required)
}
