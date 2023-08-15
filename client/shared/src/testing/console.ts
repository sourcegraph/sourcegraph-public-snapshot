import * as util from 'util'

import chalk, { type Chalk } from 'chalk'
import { identity } from 'lodash'
import type { ConsoleMessage, ConsoleMessageType, Page } from 'puppeteer'
import stringWidth from 'string-width'
import terminalSize from 'term-size'

import { asError } from '@sourcegraph/common'

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
 * Formats a console message that was logged in a Puppeteer Chrome instance for
 * output on the NodeJS terminal. Tries to mirror Chrome's console output as
 * closely as possible and makes sense.
 */
export async function formatPuppeteerConsoleMessage(page: Page, message: ConsoleMessage): Promise<string> {
    const color = colors[message.type()] ?? identity
    const icon = icons[message.type()] ?? ''
    const formattedLocation =
        'at ' +
        chalk.underline(
            [message.location().url, message.location().lineNumber, message.location().columnNumber]
                .filter(Boolean)
                .join(':')
        ) +
        ` from page ${chalk.underline(page.url())}`

    // Right-align location, like in Chrome dev tools
    const locationLine = chalk.dim(
        formattedLocation &&
            '\n' +
                (!process.env.CI
                    ? ' '.repeat(Math.max(0, terminalSize().columns - stringWidth(formattedLocation))) +
                      formattedLocation
                    : '\t' + formattedLocation)
    )

    const argumentObjects = await Promise.allSettled(
        message.args().map(async argumentHandle => {
            const json = await (
                await argumentHandle.evaluateHandle(value =>
                    JSON.stringify(value, (key, value: unknown) => {
                        // Check if value is error, because Errors are not serializable but commonly logged
                        if (value instanceof Error) {
                            return value.stack
                        }
                        return value
                    })
                )
            ).jsonValue()
            const parsed: unknown = JSON.parse(json as string)
            return parsed
        })
    )

    return [
        chalk.bold('🖥  Browser console:'),
        color(
            ...[
                message.type() !== 'log' ? chalk.bold(icon, message.type()) : '',
                ...(argumentObjects.every(result => result.status === 'rejected')
                    ? // If all arguments failed or there are no arguments, fall back to the text.
                      [message.text()]
                    : argumentObjects.map(result => {
                          if (result.status === 'rejected') {
                              return chalk.italic(`[Could not serialize: ${asError(result.reason).message}]`)
                          }
                          if (typeof result.value === 'string') {
                              return result.value
                          }
                          return util.inspect(result.value)
                      })),
                locationLine,
            ].filter(Boolean)
        ),
    ].join(' ')
}
