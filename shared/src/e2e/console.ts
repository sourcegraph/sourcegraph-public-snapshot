import { ConsoleMessage } from 'playwright'
import chalk, { Chalk } from 'chalk'
import * as util from 'util'
import terminalSize from 'term-size'
import stringWidth from 'string-width'
import { identity } from 'lodash'
import { asError } from '../util/errors'

/* Taken from type defintion of `ConsoleMessage.type` */
type ConsoleMessageType =
    | 'log'
    | 'debug'
    | 'info'
    | 'error'
    | 'warning'
    | 'dir'
    | 'dirxml'
    | 'table'
    | 'trace'
    | 'clear'
    | 'startGroup'
    | 'startGroupCollapsed'
    | 'endGroup'
    | 'assert'
    | 'profile'
    | 'profileEnd'
    | 'count'
    | 'timeEnd'

const colors: Partial<Record<ConsoleMessageType, Chalk>> = {
    error: chalk.red,
    warning: chalk.yellow,
    info: chalk.cyan,
}
const icons: Partial<Record<ConsoleMessageType, string>> = {
    error: '✖',
    warning: '⚠',
    info: 'ℹ',
}

/**
 * Formats a console message that was logged in a Puppeteer Chrome instance for output on the NodeJS terminal.
 * Tries to mirror Chrome's console output as closely as possible and makes sense.
 */
export async function formatPlaywrightConsoleMessage(message: ConsoleMessage): Promise<string> {
    const color = colors[message.type() as ConsoleMessageType] ?? identity
    const icon = icons[message.type() as ConsoleMessageType] ?? ''
    const formattedLocation =
        'at ' +
        chalk.underline(
            [message.location().url, message.location().lineNumber, message.location().columnNumber]
                .filter(Boolean)
                .join(':')
        )
    // Right-align location, like in Chrome dev tools
    const locationLine = chalk.dim(
        formattedLocation &&
            '\n' +
                (!process.env.CI
                    ? ' '.repeat(terminalSize().columns - stringWidth(formattedLocation)) + formattedLocation
                    : '\t' + formattedLocation)
    )
    return [
        chalk.bold('🖥  Browser console:'),
        color(
            ...[
                message.type() !== 'log' ? chalk.bold(icon, message.type()) : '',
                message.args().length === 0 ? message.text() : '',
                ...(
                    await Promise.all(
                        message.args().map(async argHandle => {
                            try {
                                const json = await (
                                    await argHandle.evaluateHandle(value =>
                                        JSON.stringify(value, (key, value: unknown) => {
                                            // Check if value is error, because Errors are not serializable but commonly logged
                                            if (value instanceof Error) {
                                                return value.stack
                                            }
                                            return value
                                        })
                                    )
                                ).jsonValue()
                                const parsed: unknown = JSON.parse(json)
                                return parsed
                            } catch (err) {
                                return chalk.italic(`[Could not serialize: ${asError(err).message}]`)
                            }
                        })
                    )
                ).map(value => (typeof value === 'string' ? value : util.inspect(value))),
                locationLine,
            ].filter(Boolean)
        ),
    ].join(' ')
}
