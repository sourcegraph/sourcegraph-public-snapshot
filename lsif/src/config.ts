import got from 'got'
import { readEnvInt } from './util'
import * as json5 from 'json5'
import { Logger } from 'winston'

/**
 * HTTP address for internal frontend HTTP API.
 */
const SRC_FRONTEND_INTERNAL = process.env.SRC_FRONTEND_INTERNAL || 'sourcegraph-frontend-internal'

/**
 * How long to wait before emitting error logs when polling config (in seconds).
 * This needs to be long enough to allow the frontend to fully migrate the PostgreSQL
 * database in most cases, to avoid log spam when running sourcegraph/server for the
 * first time.
 */
const DELAY_BEFORE_UNREACHABLE_LOG = readEnvInt('DELAY_BEFORE_UNREACHABLE_LOG', 15)

/**
 * How long to wait between polling config.
 */
const CONFIG_POLL_INTERVAL = 5

/**
 * A container for the current configuration data. This object should
 * be passed around and its current field should be read fresh each
 * time it is used. It will be mutable updated in the background.
 */
export interface ConfigurationContext {
    /**
     * The current configuration.
     */
    current: Configuration

    /**
     * A promise that resolves once the current configuration has been
     * successfully read for the first time.
     */
    initialized: Promise<void>
}

/**
 * Service configuration data pulled from the frontend.
 */
export interface Configuration {
    /**
     * The connection string for the Postgres database.
     */
    postgresDSN: string

    /**
     * The (ordered) URLs of the registered gitservers.
     */
    gitServers: string[]

    /**
     * The access token for lightstep tracing.
     */
    lightstepAccessToken: string

    /**
     * Whether or not to enable Jaeger.
     */
    useJaeger: boolean
}

/**
 * Create a configuration context and wait for it to initialize. If any
 * of the fields that cannot be updated while the process remains up
 * changes, it will forcibly exit the process to allow whatever orchestrator
 * is running this process restart it.
 *
 * @param logger The logger instance.
 */
export async function waitForConfiguration(logger: Logger): Promise<ConfigurationContext> {
    let oldConfiguration: Configuration | undefined
    const onUpdate = (newConfiguration: Configuration): void => {
        if (oldConfiguration !== undefined && requireRestart(oldConfiguration, newConfiguration)) {
            console.error('Detected configuration change, restarting to take effect')
            process.exit(1)
        }

        oldConfiguration = newConfiguration
    }

    const context = watchConfig(logger, onUpdate)
    await context.initialized
    return context
}

/**
 * Determine if the two configurations differ by a field that cannot be changed
 * while the process remains up and a restart would be required for the change to
 * take effect.
 *
 * @param oldConfiguration The old configuration instance.
 * @param newConfiguration The new configuration instance.
 */
function requireRestart(oldConfiguration: Configuration, newConfiguration: Configuration): boolean {
    const fields: ('postgresDSN' | 'lightstepAccessToken' | 'useJaeger')[] = [
        'postgresDSN',
        'lightstepAccessToken',
        'useJaeger',
    ]

    for (const field of fields) {
        if (oldConfiguration[field] !== newConfiguration[field]) {
            return true
        }
    }

    return false
}

/**
 * Create a configuration context instance. This will also start
 * polling the frontend server on an infinite loop in the background.
 *
 * @param logger The logger instance.
 * @param onChange The callback to invoke each time the configuration is read.
 */
function watchConfig(logger: Logger, onChange: (newConfiguration: Configuration) => void): ConfigurationContext {
    const emptyConfiguration: Configuration = {
        postgresDSN: '',
        gitServers: [],
        lightstepAccessToken: '',
        useJaeger: false,
    }

    const ctx: ConfigurationContext = {
        current: emptyConfiguration,
        initialized: new Promise<void>(resolve => {
            const onUpdate = (configuration: Configuration): void => {
                ctx.current = configuration
                onChange(configuration)
                resolve()
            }

            // Can't fail, but call catch to reduce lint noise
            updateConfiguration(logger, onUpdate).catch(() => {})
        }),
    }

    return ctx
}

/**
 * Read the configuration from the frontend on a loop. This function is async but does not
 * return any meaningful value (the returned promise neither resolves nor rejects).
 *
 * @param logger The logger instance.
 * @param onChange The callback to invoke each time the configuration is read.
 */
async function updateConfiguration(logger: Logger, onChange: (configuration: Configuration) => void): Promise<void> {
    const start = Date.now()

    while (true) {
        try {
            onChange(await loadConfiguration())
        } catch (error) {
            // Suppress log messages for errors caused by the frontend being unreachable until we've
            // given the frontend enough time to initialize (in case other services start up before
            // the frontend), to reduce log spam.
            if (Date.now() - start > DELAY_BEFORE_UNREACHABLE_LOG * 1000 || error.code !== 'ECONNREFUSED') {
                logger.error('failed to retrieve configuration from frontend', { error })
            }
        }

        // Do a jittery sleep _up to_ the config poll interval.
        const durationMs = Math.floor(Math.random() * CONFIG_POLL_INTERVAL * 1000)
        await new Promise(resolve => setTimeout(resolve, durationMs))
    }
}

/**
 * Read configuration from the frontend.
 */
async function loadConfiguration(): Promise<Configuration> {
    const url = new URL(`http://${SRC_FRONTEND_INTERNAL}/.internal/configuration`).href
    const resp = await got.post(url, { followRedirect: true })
    const payload = JSON.parse(resp.body)

    // Already parsed
    const serviceConnections = payload.ServiceConnections
    // Need to parse but must support comments + trailing commas
    const critical = json5.parse(payload.Critical)

    return {
        gitServers: serviceConnections.gitServers,
        postgresDSN: serviceConnections.postgresDSN,
        lightstepAccessToken: critical.lightstepAccessToken,
        useJaeger: critical.useJaeger || false,
    }
}
