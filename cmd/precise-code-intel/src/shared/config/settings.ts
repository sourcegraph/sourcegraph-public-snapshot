import { readEnvInt } from '../settings'

/** HTTP address for internal frontend HTTP API. */
export const SRC_FRONTEND_INTERNAL = process.env.SRC_FRONTEND_INTERNAL || 'sourcegraph-frontend-internal'

/**
 * How long to wait before emitting error logs when polling config (in seconds).
 * This needs to be long enough to allow the frontend to fully migrate the PostgreSQL
 * database in most cases, to avoid log spam when running sourcegraph/server for the
 * first time.
 */
export const DELAY_BEFORE_UNREACHABLE_LOG = readEnvInt('DELAY_BEFORE_UNREACHABLE_LOG', 15)

/**
 * How long to wait between emitting error logs when polling config (in seconds). This
 * reduces the chatter in the terminal when services are restarting.
 */
export const DELAY_BETWEEN_UNREACHABLE_LOG = readEnvInt('DELAY_BETWEEN_UNREACHABLE_LOG', 30)

/** How long to wait between polling config. */
export const CONFIG_POLL_INTERVAL = 5
