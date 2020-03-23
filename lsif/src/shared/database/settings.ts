import { readEnvInt } from '../settings'

/** How many times to try to check the current database migration version on startup. */
export const MAX_SCHEMA_CHECK_RETRIES = readEnvInt('MAX_SCHEMA_POLL_RETRIES', 60)

/** How long to wait (minimum, in seconds) between queries to check the current database migration version on startup. */
export const MIN_SCHEMA_CHECK_RETRY_TIMEOUT = readEnvInt('MIN_SCHEMA_CHECK_RETRY_TIMEOUT', 1)

/** How long to wait (maximum, in seconds) between queries to check the current database migration version on startup. */
export const MAX_SCHEMA_CHECK_RETRY_TIMEOUT = readEnvInt('MAX_SCHEMA_CHECK_RETRY_TIMEOUT', 30)

/** How many times to try to connect to Postgres on startup. */
export const MAX_CONNECTION_RETRIES = readEnvInt('MAX_CONNECTION_RETRIES', 60)

/** How long to wait (minimum, in seconds) between Postgres connection attempts. */
export const MIN_CONNECTION_RETRY_TIMEOUT = readEnvInt('MIN_CONNECTION_RETRY_TIMEOUT', 1)

/** How long to wait (maximum, in seconds) between Postgres connection attempts. */
export const MAX_CONNECTION_RETRY_TIMEOUT = readEnvInt('MAX_CONNECTION_RETRY_TIMEOUT', 30)
