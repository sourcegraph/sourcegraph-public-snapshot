import * as json5 from 'json5'
import * as settings from './settings'
import got from 'got'
import { isEqual, pick } from 'lodash'
import { Logger } from 'winston'
import delay from 'delay'

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
 * Create a configuration fetcher function and block until the first payload
 * can be read from the frontend. Continue reading the configuration from the
 * frontend in the background. If one of the fields that cannot be updated while
 * the process remains up changes, it will forcibly exit the process to allow
 * whatever orchestrator is running this process restart it.
 *
 * @param logger The logger instance.
 */
export async function waitForConfiguration(logger: Logger): Promise<() => Configuration> {
    let oldConfiguration: Configuration | undefined

    await new Promise<void>(resolve => {
        updateConfiguration(logger, configuration => {
            if (oldConfiguration !== undefined && requireRestart(oldConfiguration, configuration)) {
                logger.error('Detected configuration change, restarting to take effect')
                process.exit(1)
            }

            oldConfiguration = configuration
            resolve()
        }).catch(() => {})
    })

    // This value is guaranteed to be set by the resolution of the promise above
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
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
 * @param logger The logger instance.
 * @param onChange The callback to invoke each time the configuration is read.
 */
async function updateConfiguration(logger: Logger, onChange: (configuration: Configuration) => void): Promise<never> {
    const start = Date.now()
    while (true) {
        try {
            onChange(await loadConfiguration())
        } catch (error) {
            // Suppress log messages for errors caused by the frontend being unreachable until we've
            // given the frontend enough time to initialize (in case other services start up before
            // the frontend), to reduce log spam.
            if (Date.now() - start > settings.DELAY_BEFORE_UNREACHABLE_LOG * 1000 || error.code !== 'ECONNREFUSED') {
                logger.error('Failed to retrieve configuration from frontend', { error })
            }
        }

        // Do a jittery sleep _up to_ the config poll interval.
        const durationMs = Math.random() * settings.CONFIG_POLL_INTERVAL * 1000
        await delay(durationMs)
    }
}

/**
 * Read configuration from the frontend.
 */
async function loadConfiguration(): Promise<Configuration> {
    const url = new URL(`http://${settings.SRC_FRONTEND_INTERNAL}/.internal/configuration`).href
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
