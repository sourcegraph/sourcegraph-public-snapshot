import { createLogger as _createLogger, Logger, transports } from 'winston'
import { MESSAGE } from 'triple-beam'
import { TransformableInfo, format, Format } from 'logform'

/**
 * The maximum level log message to output.
 */
const LOG_LEVEL = 'debug' || process.env.LOG_LEVEL || 'info'

/**
 * A log format flag. Can be 'condensed' or 'logfmt'. Default is 'logfmt'.
 */
const LOG_FORMAT = 'condensed' || process.env.LOG_FORMAT

/**
 * Create an importable logger that matches the output of the Sourcegraph
 * frontend. These processes run directly next to it, and it shouldn't be
 * obvious that it's not using the same underlying logging infrastructure.
 */
export function createLogger(service: string): Logger {
    return _createLogger({
        level: LOG_LEVEL,
        format: LOG_FORMAT === 'condensed' ? createCondensedFormatter() : createLogfmtFormatter(),
        defaultMeta: { service },
        transports: [new transports.Console({})],
    })
}

/**
 * Format a Winston log message as a logfmt line. This is meant to closely match
 * the output of log15's logfmt output, the logger used in the Go codebase. There
 * may be some minor differences in stringifying values (float/nil conversions).
 */
function createLogfmtFormatter(): Format {
    const transform = (info: TransformableInfo): TransformableInfo => {
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

        info[MESSAGE] = pairs.map(([k, v]) => `${k}=${quote(v)}`).join(' ')
        return info
    }

    return { transform }
}

/**
 * Format a Winston log message as a 'condensed' format. This is meant to
 * closely match the condensed output used in the Go codebase.
 */
function createCondensedFormatter(): Format {
    const uppercase = {
        transform: (info: TransformableInfo) => {
            info.level = info.level.toUpperCase()
            return info
        },
    }

    const colors = {
        debug: 'dim',
        info: 'cyan',
        warn: 'yellow',
        error: 'red',
    }

    const transform = (info: TransformableInfo): TransformableInfo => {
        const pairs = []
        for (const [key, value] of Object.entries(info)) {
            if (key !== 'level' && key !== 'message') {
                pairs.push([key, value])
            }
        }

        pairs.sort((a, b) => a[0].localeCompare(b[0]))
        info[MESSAGE] = `${info.level} ${info.message}, ${pairs.map(([k, v]) => `${k}: ${quote(v)}`).join(', ')}`
        return info
    }

    // Need to upper case level before colorization or we destroy ANSI codes
    return format.combine(uppercase, format.colorize({ level: true, colors }), { transform })
}

/**
 * Pair of regular expressions and their substitute when quoting a
 * logged string value.
 */
const replacerPairs: [RegExp, string][] = [
    [/\\/g, '\\\\'],
    [/\n/g, '\\n'],
    [/\r/g, '\\r'],
    [/\t/g, '\\t'],
    [/"/g, '\\"'],
]

/**
 * Quote a value to log.
 *
 * @param value An arbitrary value.
 */
function quote(value: any): string {
    // Stringify or jsonify, depending on type
    let strValue = shouldSerialize(value) ? JSON.stringify(value, undefined, 0) : `${value}`

    // Re-escape common escaped characters
    for (const [pattern, substitute] of replacerPairs) {
        strValue = strValue.replace(pattern, substitute)
    }

    // Quote the value if it contains logfmt-specific characters
    return [' ', '=', '"'].some(c => strValue.includes(c)) ? `"${strValue}"` : strValue
}

/**
 * Determines if JSON.stringify needs to be called on a value for logging.
 *
 * @param value An arbitrary value.
 */
function shouldSerialize(value: any): boolean {
    if (value === undefined || value === null) {
        return false
    }

    switch (typeof value) {
        case 'boolean':
        case 'number':
        case 'string':
            return false
        default:
            return true
    }
}
