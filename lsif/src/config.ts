import got from 'got'
import { readEnvInt } from './util'
import * as json5 from 'json5'
import { isEqual, pick } from 'lodash'

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
     * The project name for lightstep tracing.
     */
    lightstepProject: string

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
 * A function that returns the current configuration.
 */
export type ConfigurationFetcher = () => Configuration

/**
 * Create a configuration fetcher function and block until the first payload
 * can be read from teh frontend. Continue reading the configuration from the
 * frontend in the background. If one of the fields that cannot be updated while
 * the process remains up changes, it will forcibly exit the process to allow
 * whatever orchestrator is running this process restart it.
 */
export async function waitForConfiguration(): Promise<ConfigurationFetcher> {
    let oldConfiguration: Configuration | undefined

    await new Promise<void>(resolve => {
        updateConfiguration(configuration => {
            if (oldConfiguration !== undefined && requireRestart(oldConfiguration, configuration)) {
                console.error('Detected configuration change, restarting to take effect')
                process.exit(1)
            }

            oldConfiguration = configuration
            resolve()
        }).catch(() => {})
    })

    return () => oldConfiguration!
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
    const fields = ['postgresDSN', 'lightstepProject', 'lightstepAccessToken', 'useJaeger']

    return !isEqual(pick(oldConfiguration, fields), pick(newConfiguration, fields))
}

/**
 * Read the configuration from the frontend on a loop. This function is async but does not
 * return any meaningful value (the returned promise neither resolves nor rejects).
 *
 * @param onChange The callback to invoke each time the configuration is read.
 */
async function updateConfiguration(onChange: (configuration: Configuration) => void): Promise<never> {
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
        const durationMs = Math.random() * CONFIG_POLL_INTERVAL * 1000
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
        lightstepProject: critical.lightstepProject,
        lightstepAccessToken: critical.lightstepAccessToken,
        useJaeger: critical.useJaeger || false,
    }
}
