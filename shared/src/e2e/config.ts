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
    envVar?: string
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
            const info = []
            if (field.envVar) {
                info.push(`environment variable: ${field.envVar}`)
            }
            if (field.description) {
                info.push(`description: ${field.description}`)
            }
            if (field.defaultValue) {
                info.push(`default value: ${field.defaultValue}`)
            }
            return `(${info.join(', ')})`
        }
        // const fieldInfo = k => `- ${k}`
        throw new Error(`FAIL: Required config was not provided. The following keys were missing:

${missingKeys.map(k => `- ${k} ${fieldInfo(k)}`).join('\n')}

Please set the appropriate environment variables or add these entries to the config file
specified by the environment variable CONFIG_FILE`)
    }

    return pick(config, required)
}
