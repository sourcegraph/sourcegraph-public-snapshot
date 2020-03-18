import { readEnvInt } from '../settings'

/** How many times to try to check the current database migration version on startup. */
export const MAX_SCHEMA_POLL_RETRIES = readEnvInt('MAX_SCHEMA_POLL_RETRIES', 60)

/** How long to wait (in seconds) between queries to check the current database migration version on startup. */
export const SCHEMA_POLL_INTERVAL = readEnvInt('SCHEMA_POLL_INTERVAL', 5)

/** How many times to try to connect to Postgres on startup. */
export const MAX_CONNECTION_RETRIES = readEnvInt('MAX_CONNECTION_RETRIES', 60)

/** How long to wait (in seconds) between Postgres connection attempts. */
export const CONNECTION_RETRY_INTERVAL = readEnvInt('CONNECTION_RETRY_INTERVAL', 5)
