import { createLogger as _createLogger, Logger, transports } from 'winston'
import { format, TransformableInfo } from 'logform'
import { inspect } from 'util'
import { MESSAGE } from 'triple-beam'

/**
 * Return a sanitized log level.
 *
 * @param value The raw log level.
 */
function toLogLevel(value: string): 'debug' | 'info' | 'warn' | 'error' {
    if (value === 'debug' || value === 'info' || value === 'warn' || value === 'error') {
        return value
    }

    return 'info'
}

/**
 * The maximum level log message to output.
 */
const LOG_LEVEL: 'debug' | 'info' | 'warn' | 'error' = toLogLevel((process.env.LOG_LEVEL || 'info').toLowerCase())

/**
 * Create a structured logger.
 */
export function createLogger(service: string): Logger {
    const formatTransformer = (info: TransformableInfo): TransformableInfo => {
        const attributes: { [name: string]: unknown } = {}
        for (const [key, value] of Object.entries(info)) {
            if (key !== 'level' && key !== 'message') {
                attributes[key] = value
            }
        }

        info[MESSAGE] = `${info.level} ${info.message} ${inspect(attributes)}`
        return info
    }

    const uppercaseTransformer = (info: TransformableInfo): TransformableInfo => {
        info.level = info.level.toUpperCase()
        return info
    }

    const colors = {
        debug: 'dim',
        info: 'cyan',
        warn: 'yellow',
        error: 'red',
    }

    return _createLogger({
        level: LOG_LEVEL,
        // Need to upper case level before colorization or we destroy ANSI codes
        format: format.combine({ transform: uppercaseTransformer }, format.colorize({ level: true, colors }), {
            transform: formatTransformer,
        }),
        defaultMeta: { service },
        transports: [new transports.Console({})],
    })
}

/**
 * Creates a silent logger.
 */
export function createSilentLogger(): Logger {
    return _createLogger({ silent: true })
}

/**
 * Log the beginning, end, and exception of an operation.
 *
 * @param name The log message to output.
 * @param logger The logger instance.
 * @param f The operation to perform.
 */
export async function logCall<T>(name: string, logger: Logger, f: () => Promise<T> | T): Promise<T> {
    const timer = logger.startTimer()
    logger.debug(`${name}: starting`)

    try {
        const value = await f()
        timer.done({ message: `${name}: finished`, level: 'debug' })
        return value
    } catch (error) {
        timer.done({ message: `${name}: failed`, level: 'error', error })
        throw error
    }
}
