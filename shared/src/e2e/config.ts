import * as fs from 'fs'
import { pick } from 'lodash'

/**
 * Defines configuration for e2e tests. This is as-yet incomplete as some config
 * depended on by other modules is not included here.
 */
export interface Config {
    sudoToken: string
    username: string
    gitHubToken: string
    sourcegraphBaseUrl: string
}

export interface ConfigField {
    envVar: string
    description?: string
    defaultValue?: string
}

const configFields: { [K in keyof Config]: ConfigField } = {
    sudoToken: {
        envVar: 'SOURCEGRAPH_SUDO_TOKEN',
    },
    username: {
        envVar: 'SOURCEGRAPH_USERNAME',
    },
    gitHubToken: {
        envVar: 'GITHUB_TOKEN',
    },
    sourcegraphBaseUrl: {
        envVar: 'SOURCEGRAPH_BASE_URL',
        defaultValue: 'http://localhost:3080',
    },
}

/**
 * Reads e2e config from appropriate inputs: the config file defined by $CONFIG_FILE
 * and environment variables. The caller should specify the config fields that it
 * depends on. This method reads the config synchronously from disk.
 */
export function getConfig<T extends keyof Config>(required: T[]): Pick<Config, T> {
    const configFile = process.env.CONFIG_FILE
    // eslint-disable-next-line no-sync
    const config = configFile ? JSON.parse(fs.readFileSync(configFile, 'utf-8')) : {}

    // Set defaults and read env vars
    for (const [fieldName, field] of Object.entries(configFields)) {
        if (field.defaultValue && config[fieldName] === undefined) {
            config[fieldName] = field.defaultValue
        }
        if (field.envVar && process.env[field.envVar]) {
            config[fieldName] = process.env[field.envVar]
        }
    }

    // Check required fields for type safety
    const missingKeys = required.filter(key => !config[key])
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
