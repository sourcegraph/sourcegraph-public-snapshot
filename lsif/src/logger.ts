import winston from 'winston'
import { TransformableInfo } from 'logform'
import { MESSAGE } from 'triple-beam'

/**
 * The maximum level log message to output.
 */
const LOG_LEVEL = process.env.LOG_LEVEL || 'info'

/**
 * A log format flag. Can be 'condensed' or 'logfmt'. uses 'logfmt'
 * by default.
 */
const LOG_FORMAT = process.env.LOG_FORMAT

/**
 * Whether or not to disable colorization in the condensed formatter.
 */
const NO_COLOR = !!process.env.NO_COLOR

/**
 * A map of log levels to colors. Used in the condensed formatter when
 * NO_COLOR is not enabled.
 */
const colors: { [k: string]: (text: string) => string } = {
    error: NO_COLOR ? text => text : text => `\x1b[31m${text}\x1b[0m`,
    warn: NO_COLOR ? text => text : text => `\x1b[33m${text}\x1b[0m`,
    info: NO_COLOR ? text => text : text => `\x1b[36m${text}\x1b[0m`,
    debug: NO_COLOR ? text => text : text => `\x1b[2m${text}\x1b[0m`,
}

/**
 * Format a Winston log message as a 'condensed' format. This is meant to
 * closely match the condensed output used in the Go codebase.
 */
function condensedFormat(info: TransformableInfo, opts?: any): TransformableInfo {
    const pairs = []
    for (const [key, value] of Object.entries(info)) {
        if (key !== 'level' && key !== 'message') {
            pairs.push([key, value])
        }
    }

    pairs.sort((a, b) => a[0].localeCompare(b[0]))
    const level = colors[info.level](info.level.toUpperCase())
    info[MESSAGE] = `${level} ${info.message}, ${pairs.map(([k, v]) => `${k}: ${v}`).join(', ')}`
    return info
}

/**
 * Format a Winston log message as a logfmt line. This is meant to closely match
 * the output of log15's logfmt output, the logger used in the Go codebase. There
 * may be some minor differences in stringifying values (float/nil conversions).
 */
function logfmtFormat(info: TransformableInfo, opts?: any): TransformableInfo {
    const pairs = []
    pairs.push(['t', info.timestamp ? info.timestamp : new Date().toISOString()])
    pairs.push(['lvl', info.level])
    pairs.push(['msg', info.message])

    const additionalPairs = []
    for (const [key, value] of Object.entries(info)) {
        if (key !== 'timestamp' && key !== 'level' && key !== 'message') {
            additionalPairs.push([key, value])
        }
    }

    additionalPairs.sort((a, b) => a[0].localeCompare(b[0]))
    pairs.push(...additionalPairs)

    const quote = (value: any): string => {
        let strValue = `${value}`
        const needsQuotes = [' ', '=', '"'].some(c => strValue.includes(c))

        strValue = strValue.replace(/\\/g, '\\\\')
        strValue = strValue.replace(/\n/g, '\\n')
        strValue = strValue.replace(/\r/g, '\\r')
        strValue = strValue.replace(/\t/g, '\\t')
        strValue = strValue.replace(/"/g, '\\"')

        return needsQuotes ? `"${strValue}"` : strValue
    }

    info[MESSAGE] = pairs
        .map(([k, v]) => [k, quote(v)])
        .map(([k, v]) => `${k}=${v}`)
        .join(' ')

    return info
}

/**
 * Wrap the formatter function in a class acceptable to Winston.
 */
class Formatter {
    public transform = LOG_FORMAT === 'condensed' ? condensedFormat : logfmtFormat
}

/**
 * An importable logger. This must be initialized via `initLogger` at
 * application startup.
 */
export let logger!: winston.Logger

/**
 * Create an importable logger that matches the output of the Sourcegraph
 * frontend. These processes run directly next to it, and it shouldn't be
 * obvious that it's not using the same underlying logging infrastructure.
 */
export function initLogger(service: string): void {
    logger = winston.createLogger({
        level: LOG_LEVEL,
        format: new Formatter(),
        defaultMeta: { service },
        transports: [new winston.transports.Console({})],
    })
}
