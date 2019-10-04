import got from 'got'
import { readEnvInt } from './util'

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
}

/**
 * Create a configuration context instance. This will also start
 * polling the frontend server on an infinite loop in the background.
 *
 * @param onChange The callback to invoke each time the configuration is read.
 */
export function watchConfig(onChange: (newConfiguration: Configuration) => void): ConfigurationContext {
    const emptyConfiguration: Configuration = {
        postgresDSN: '',
        gitServers: [],
    }

    const ctx: ConfigurationContext = {
        current: emptyConfiguration,
        initialized: new Promise<void>(resolve => {
            updateConfiguration(configuration => {
                ctx.current = configuration
                onChange(configuration)
                resolve()
            }).catch(() => {})
        }),
    }

    return ctx
}

/**
 * Read the configuration from the frontend on a loop. This function is async but does not
 * return any meaningful value (the returned promise neither resolves nor rejects).
 *
 * @param onChange The callback to invoke each time the configuration is read.
 */
async function updateConfiguration(onChange: (configuration: Configuration) => void): Promise<void> {
    const start = Date.now()
    while (true) {
        try {
            onChange(await loadConfiguration())
        } catch (error) {
            // Suppress log messages for errors caused by the frontend being unreachable until we've
            // given the frontend enough time to initialize (in case other services start up before
            // the frontend), to reduce log spam.
            if (Date.now() - start > DELAY_BEFORE_UNREACHABLE_LOG * 1000 || error.code !== 'ECONNREFUSED') {
                console.error(error)
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

    return {
        gitServers: payload.ServiceConnections.gitServers,
        postgresDSN: payload.ServiceConnections.postgresDSN,
    }
}
